package helm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/PastureStack/kubectl-service/cli"
)

func useHelm4Backend(t *testing.T) {
	t.Helper()
	if err := ConfigureBackend(Helm4BackendName); err != nil {
		t.Fatalf("configure Helm 4 backend: %v", err)
	}
	t.Cleanup(func() {
		if err := ConfigureBackend(LegacyHelm2BackendName); err != nil {
			t.Fatalf("restore legacy backend: %v", err)
		}
	})
}

func TestHelm4InstallPreservesClientSideApplyContract(t *testing.T) {
	useHelm4Backend(t)
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if cmd != helm4Command {
			return cli.Output{ExitCode: 1, StdErr: fmt.Sprintf("unexpected command %s", cmd)}
		}
		return cli.Output{StdOut: "install notes"}
	})

	notes, err := InstallHelmStack(minimalChartStack("demo", "team-a"))
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if notes != "install notes" || len(*calls) != 2 {
		t.Fatalf("unexpected result notes=%q calls=%#v", notes, *calls)
	}
	if !reflect.DeepEqual((*calls)[0].args[:2], []string{"dependency", "update"}) {
		t.Fatalf("unexpected dependency call %#v", (*calls)[0])
	}
	installArgs := (*calls)[1].args
	if len(installArgs) != 7 || installArgs[0] != "install" || installArgs[1] != "demo" ||
		installArgs[3] != "--namespace" || installArgs[4] != "team-a" ||
		installArgs[5] != "--server-side=false" || installArgs[6] != "--color=never" {
		t.Fatalf("unexpected Helm 4 install args %#v", installArgs)
	}
	if (*calls)[0].args[2] != installArgs[2] {
		t.Fatalf("dependency and install chart paths differ: %#v", *calls)
	}
}

func TestHelm4InstallStopsAfterDependencyFailure(t *testing.T) {
	useHelm4Backend(t)
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if cmd != helm4Command || len(args) < 2 || args[0] != "dependency" || args[1] != "update" {
			return cli.Output{ExitCode: 1, StdErr: "unexpected command"}
		}
		return cli.Output{ExitCode: 1, StdErr: "dependency failed"}
	})

	if _, err := InstallHelmStack(minimalChartStack("demo", "team-a")); err == nil || !strings.Contains(err.Error(), "dependency failed") {
		t.Fatalf("unexpected dependency failure: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("install continued after dependency failure: %#v", *calls)
	}
}

func TestHelm4UpgradeRollbackAndListContracts(t *testing.T) {
	useHelm4Backend(t)
	listJSON := `[{"name":"demo","namespace":"team-a","revision":"3","updated":"2026-08-07T08:00:00Z","status":"deployed","chart":"demo-0.1.0","app_version":"1.0"}]`
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if len(args) > 0 && args[0] == "list" {
			return cli.Output{StdOut: listJSON}
		}
		return cli.Output{StdOut: "upgrade notes"}
	})

	notes, err := UpgradeHelmStack(minimalChartStack("demo", "team-a"))
	if err != nil || notes != "upgrade notes" {
		t.Fatalf("upgrade result notes=%q err=%v", notes, err)
	}
	if err := RollbackHelmStack(&Stack{Name: "demo", Namespace: "team-a"}); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	releases, err := ListReleases()
	if err != nil || len(releases) != 1 {
		t.Fatalf("list result releases=%#v err=%v", releases, err)
	}
	if releases[0].Namespace != "team-a" || releases[0].Revision != 3 || releases[0].Status != "DEPLOYED" {
		t.Fatalf("unexpected parsed release %#v", releases[0])
	}

	upgradeArgs := (*calls)[1].args
	if upgradeArgs[0] != "upgrade" || upgradeArgs[5] != "--server-side=false" {
		t.Fatalf("unexpected upgrade args %#v", upgradeArgs)
	}
	rollbackArgs := (*calls)[2].args
	if !reflect.DeepEqual(rollbackArgs, []string{"rollback", "demo", "0", "--namespace", "team-a", "--server-side=false", "--color=never"}) {
		t.Fatalf("unexpected rollback args %#v", rollbackArgs)
	}
	listArgs := (*calls)[3].args
	if !reflect.DeepEqual(listArgs, []string{"list", "--all-namespaces", "--output", "json", "--max", "10000", "--color=never"}) {
		t.Fatalf("unexpected list args %#v", listArgs)
	}
}

func TestHelm4DeleteRequiresUnambiguousNamespace(t *testing.T) {
	useHelm4Backend(t)
	listJSON := `[
		{"name":"demo","namespace":"team-a","revision":1,"updated":"2026-08-07T08:00:00Z","status":"deployed","chart":"demo-0.1.0"},
		{"name":"demo","namespace":"team-b","revision":2,"updated":"2026-08-07T08:01:00Z","status":"failed","chart":"demo-0.1.0"}
	]`
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if len(args) > 0 && args[0] == "list" {
			return cli.Output{StdOut: listJSON}
		}
		return cli.Output{}
	})

	err := DeleteHelmStack(&Stack{Name: "demo"})
	if err == nil || !strings.Contains(err.Error(), "ambiguous across namespaces") {
		t.Fatalf("unexpected ambiguous delete result: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("delete should stop after the ambiguous list: %#v", *calls)
	}
}

func TestHelm4DeleteUsesExactNamespaceAndThenCleansIt(t *testing.T) {
	useHelm4Backend(t)
	listJSON := `[{"name":"demo","namespace":"team-a","revision":1,"updated":"2026-08-07 08:00:00 +0000 UTC","status":"failed","chart":"demo-0.1.0"}]`
	namespaces := []string{}
	originalDeleteNamespace := deleteKubernetesNamespace
	deleteKubernetesNamespace = func(namespace string) error {
		namespaces = append(namespaces, namespace)
		return nil
	}
	t.Cleanup(func() { deleteKubernetesNamespace = originalDeleteNamespace })
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if len(args) > 0 && args[0] == "list" {
			return cli.Output{StdOut: listJSON}
		}
		return cli.Output{}
	})

	if err := DeleteHelmStack(&Stack{Name: "demo", Namespace: "team-a"}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(*calls) != 2 || !reflect.DeepEqual((*calls)[1].args, []string{"uninstall", "demo", "--namespace", "team-a", "--color=never"}) {
		t.Fatalf("unexpected delete calls %#v", *calls)
	}
	if !reflect.DeepEqual(namespaces, []string{"team-a"}) {
		t.Fatalf("unexpected namespace cleanup %#v", namespaces)
	}
}

func TestHelm4DeleteWithoutCallerNamespaceUsesDiscoveredNamespaceWithoutRemovingIt(t *testing.T) {
	useHelm4Backend(t)
	originalDeleteNamespace := deleteKubernetesNamespace
	deleteKubernetesNamespace = func(string) error {
		t.Fatal("discovered namespace must not be removed when the caller omitted it")
		return nil
	}
	t.Cleanup(func() { deleteKubernetesNamespace = originalDeleteNamespace })
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if len(args) > 0 && args[0] == "list" {
			return cli.Output{StdOut: `[{"name":"demo","namespace":"team-a","revision":1,"updated":"","status":"deployed","chart":"demo-0.1.0"}]`}
		}
		return cli.Output{}
	})

	if err := DeleteHelmStack(&Stack{Name: "demo"}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(*calls) != 2 || !reflect.DeepEqual((*calls)[1].args, []string{"uninstall", "demo", "--namespace", "team-a", "--color=never"}) {
		t.Fatalf("unexpected delete calls %#v", *calls)
	}
}

func TestHelm4DeleteMissingReleaseIsIdempotent(t *testing.T) {
	useHelm4Backend(t)
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		return cli.Output{StdOut: `[]`}
	})

	if err := DeleteHelmStack(&Stack{Name: "missing", Namespace: "team-a"}); err != nil {
		t.Fatalf("idempotent delete: %v", err)
	}
	if len(*calls) != 1 || (*calls)[0].args[0] != "list" {
		t.Fatalf("missing release triggered extra commands: %#v", *calls)
	}
}

func TestHelm4ListRejectsMalformedAndPotentiallyTruncatedOutput(t *testing.T) {
	useHelm4Backend(t)
	withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		return cli.Output{StdOut: `[{"name":"demo","namespace":"default","revision":"bad","updated":"2026-08-07T08:00:00Z","status":"deployed","chart":"demo"}]`}
	})
	if _, err := ListReleases(); err == nil || !strings.Contains(err.Error(), "revision") {
		t.Fatalf("malformed revision returned %v", err)
	}
}

func TestHelm4ListRejectsMalformedIdentityTimeAndJSON(t *testing.T) {
	useHelm4Backend(t)
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{name: "empty name", output: `[{"name":"","namespace":"default","revision":1,"updated":"","status":"deployed","chart":"demo"}]`, want: "name is empty"},
		{name: "empty namespace", output: `[{"name":"demo","namespace":"","revision":1,"updated":"","status":"deployed","chart":"demo"}]`, want: "namespace is empty"},
		{name: "zero revision", output: `[{"name":"demo","namespace":"default","revision":0,"updated":"","status":"deployed","chart":"demo"}]`, want: "positive"},
		{name: "bad time", output: `[{"name":"demo","namespace":"default","revision":1,"updated":"not-a-time","status":"deployed","chart":"demo"}]`, want: "timestamp"},
		{name: "bad JSON", output: `[`, want: "release list"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withFakeRunner(t, func(cmd string, args ...string) cli.Output {
				return cli.Output{StdOut: test.output}
			})
			if _, err := ListReleases(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("malformed output returned %v", err)
			}
		})
	}
}

func TestHelm4ListRejectsCommandFailureAndSafetyLimit(t *testing.T) {
	useHelm4Backend(t)
	t.Run("command failure", func(t *testing.T) {
		withFakeRunner(t, func(cmd string, args ...string) cli.Output {
			return cli.Output{ExitCode: 1, StdErr: "list failed"}
		})
		_, err := ListReleases()
		var execErr *cli.ErrExec
		if !errors.As(err, &execErr) {
			t.Fatalf("command failure type = %T, want *cli.ErrExec: %v", err, err)
		}
	})

	t.Run("safety limit", func(t *testing.T) {
		item := `{"name":"demo","namespace":"default","revision":1,"updated":"","status":"deployed","chart":"demo"}`
		output := "[" + strings.Repeat(item+",", helm4ListLimit-1) + item + "]"
		withFakeRunner(t, func(cmd string, args ...string) cli.Output {
			return cli.Output{StdOut: output}
		})
		if _, err := ListReleases(); err == nil || !strings.Contains(err.Error(), "safety limit") {
			t.Fatalf("safety limit returned %v", err)
		}
	})
}
