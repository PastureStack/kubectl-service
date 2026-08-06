package helm

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/PastureStack/kubectl-service/cli"
)

func minimalChartStack(name, namespace string) *Stack {
	return &Stack{
		Name:      name,
		Namespace: namespace,
		Files: map[string]string{
			"Chart.yaml": "name: demo\nversion: 0.1.0\n",
		},
	}
}

func helm2ReleaseTable(name, status string) string {
	return strings.Join([]string{
		"NAME REVISION UPDATED STATUS CHART",
		fmt.Sprintf("%s 1 Mon Jan 2 15:04:05 2006 %s demo-0.1.0", name, status),
		"",
	}, "\n")
}

func TestUpgradeHelmStackUsesHelm2UpgradeContract(t *testing.T) {
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if cmd != "helm" {
			return cli.Output{ExitCode: 1, StdErr: fmt.Sprintf("unexpected cmd %s", cmd)}
		}
		return cli.Output{StdOut: "upgrade notes"}
	})

	notes, err := UpgradeHelmStack(minimalChartStack("demo", "demo-ns"))
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if notes != "upgrade notes" {
		t.Fatalf("unexpected notes %q", notes)
	}
	if len(*calls) != 2 {
		t.Fatalf("expected dependency update and upgrade calls, got %#v", *calls)
	}

	depArgs := (*calls)[0].args
	if len(depArgs) != 3 || !reflect.DeepEqual(depArgs[:2], []string{"dependency", "update"}) {
		t.Fatalf("unexpected dependency update args %#v", depArgs)
	}

	upgradeArgs := (*calls)[1].args
	if len(upgradeArgs) != 5 {
		t.Fatalf("unexpected upgrade arg count %#v", upgradeArgs)
	}
	if !reflect.DeepEqual(upgradeArgs[:4], []string{"upgrade", "--namespace", "demo-ns", "demo"}) {
		t.Fatalf("unexpected Helm 2 upgrade args %#v", upgradeArgs)
	}
	if depArgs[2] == "" || upgradeArgs[4] == "" || depArgs[2] != upgradeArgs[4] {
		t.Fatalf("expected upgrade and dependency update to use same chart path, dep=%q upgrade=%q", depArgs[2], upgradeArgs[4])
	}
}

func TestInstallHelmStackStopsWhenDependencyUpdateFails(t *testing.T) {
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if reflect.DeepEqual(args[:2], []string{"dependency", "update"}) {
			return cli.Output{ExitCode: 1, StdErr: "dependency failed"}
		}
		return cli.Output{StdOut: "should not install"}
	})

	_, err := InstallHelmStack(minimalChartStack("demo", "demo-ns"))
	if err == nil || !strings.Contains(err.Error(), "dependency failed") {
		t.Fatalf("expected dependency update error, got %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("expected only dependency update call, got %#v", *calls)
	}
}

func TestDeleteHelmStackUsesHelm2DeleteThenNamespaceCleanup(t *testing.T) {
	namespaces := []string{}
	originalDeleteNamespace := deleteKubernetesNamespace
	deleteKubernetesNamespace = func(namespace string) error {
		namespaces = append(namespaces, namespace)
		return nil
	}
	t.Cleanup(func() {
		deleteKubernetesNamespace = originalDeleteNamespace
	})

	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		switch {
		case reflect.DeepEqual(args, []string{"ls"}):
			return cli.Output{StdOut: helm2ReleaseTable("demo", "DEPLOYED")}
		case reflect.DeepEqual(args, []string{"delete", "demo"}):
			return cli.Output{}
		default:
			return cli.Output{ExitCode: 1, StdErr: fmt.Sprintf("unexpected args %#v", args)}
		}
	})

	if err := DeleteHelmStack(&Stack{Name: "demo", Namespace: "demo-ns"}); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if len(*calls) != 2 {
		t.Fatalf("expected helm ls and delete calls, got %#v", *calls)
	}
	if len(namespaces) != 1 || namespaces[0] != "demo-ns" {
		t.Fatalf("unexpected namespace cleanup calls %#v", namespaces)
	}
}

func TestDeleteBrokenHelmReleaseSuppressesDeleteErrorAndSkipsNamespaceCleanup(t *testing.T) {
	originalDeleteNamespace := deleteKubernetesNamespace
	deleteKubernetesNamespace = func(string) error {
		t.Fatal("namespace cleanup should not run for broken releases")
		return nil
	}
	t.Cleanup(func() {
		deleteKubernetesNamespace = originalDeleteNamespace
	})

	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		switch {
		case reflect.DeepEqual(args, []string{"ls"}):
			return cli.Output{StdOut: helm2ReleaseTable("demo", "FAILED")}
		case reflect.DeepEqual(args, []string{"delete", "demo"}):
			return cli.Output{ExitCode: 1, StdErr: "delete failed"}
		default:
			return cli.Output{ExitCode: 1, StdErr: fmt.Sprintf("unexpected args %#v", args)}
		}
	})

	if err := DeleteHelmStack(&Stack{Name: "demo", Namespace: "demo-ns"}); err != nil {
		t.Fatalf("broken release delete should be non-fatal, got %v", err)
	}
	if len(*calls) != 2 {
		t.Fatalf("expected helm ls and delete calls, got %#v", *calls)
	}
}
