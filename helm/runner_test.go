package helm

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/PastureStack/kubectl-service/cli"
)

type commandCall struct {
	cmd  string
	args []string
}

func withFakeRunner(t *testing.T, fn func(cmd string, args ...string) cli.Output) *[]commandCall {
	t.Helper()

	old := runCommand
	calls := []commandCall{}
	runCommand = func(cmd string, args ...string) cli.Output {
		copiedArgs := append([]string(nil), args...)
		calls = append(calls, commandCall{cmd: cmd, args: copiedArgs})
		return fn(cmd, args...)
	}
	t.Cleanup(func() {
		runCommand = old
	})

	return &calls
}

func TestInstallHelmStackUsesHelm2InstallContract(t *testing.T) {
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if cmd != "helm" {
			return cli.Output{ExitCode: 1, StdErr: fmt.Sprintf("unexpected cmd %s", cmd)}
		}
		return cli.Output{StdOut: "install notes"}
	})

	notes, err := InstallHelmStack(&Stack{
		Name:      "demo",
		Namespace: "demo-ns",
		Files: map[string]string{
			"Chart.yaml": "name: demo\nversion: 0.1.0\n",
		},
	})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if notes != "install notes" {
		t.Fatalf("unexpected notes %q", notes)
	}
	if len(*calls) != 2 {
		t.Fatalf("expected dependency update and install calls, got %#v", *calls)
	}

	depArgs := (*calls)[0].args
	if len(depArgs) != 3 || !reflect.DeepEqual(depArgs[:2], []string{"dependency", "update"}) {
		t.Fatalf("unexpected dependency update args %#v", depArgs)
	}

	installArgs := (*calls)[1].args
	if len(installArgs) != 6 {
		t.Fatalf("unexpected install arg count %#v", installArgs)
	}
	if !reflect.DeepEqual(installArgs[:5], []string{"install", "--namespace", "demo-ns", "--name", "demo"}) {
		t.Fatalf("unexpected Helm 2 install args %#v", installArgs)
	}
	if depArgs[2] == "" || installArgs[5] == "" || depArgs[2] != installArgs[5] {
		t.Fatalf("expected install and dependency update to use same chart path, dep=%q install=%q", depArgs[2], installArgs[5])
	}
}

func TestRollbackHelmStackUsesHelm2RollbackContract(t *testing.T) {
	calls := withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		return cli.Output{}
	})

	if err := RollbackHelmStack(&Stack{Name: "demo"}); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("expected one rollback call, got %#v", *calls)
	}
	if !reflect.DeepEqual((*calls)[0].args, []string{"rollback", "demo", "0"}) {
		t.Fatalf("unexpected rollback args %#v", (*calls)[0].args)
	}
}

func TestListReleasesParsesHelm2TableOutput(t *testing.T) {
	withFakeRunner(t, func(cmd string, args ...string) cli.Output {
		if !reflect.DeepEqual(args, []string{"ls"}) {
			return cli.Output{ExitCode: 1, StdErr: fmt.Sprintf("unexpected args %#v", args)}
		}
		return cli.Output{StdOut: strings.Join([]string{
			"NAME REVISION UPDATED STATUS CHART",
			"demo 1 Mon Jan 2 15:04:05 2006 DEPLOYED demo-0.1.0",
			"",
		}, "\n")}
	})

	releases, err := ListReleases()
	if err != nil {
		t.Fatalf("list releases failed: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected one release, got %#v", releases)
	}
	if releases[0].Name != "demo" || releases[0].Revision != 1 || releases[0].Status != "DEPLOYED" || releases[0].Chart != "demo-0.1.0" {
		t.Fatalf("unexpected release parsed: %#v", releases[0])
	}
}

func TestCollapseContiguousSpacesHasNoSharedState(t *testing.T) {
	if got := collapseContiguousSpaces("a\t\tb   c"); got != "a b c" {
		t.Fatalf("unexpected first collapse result %q", got)
	}
	if got := collapseContiguousSpaces("  d"); got != " d" {
		t.Fatalf("unexpected second collapse result %q", got)
	}
}
