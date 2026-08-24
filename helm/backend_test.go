package helm

import (
	"testing"

	"github.com/PastureStack/kubectl-service/cli"
)

func TestActiveBackendIsHelm4(t *testing.T) {
	if got := ActiveBackendName(); got != Helm4BackendName {
		t.Fatalf("unexpected active backend %q", got)
	}
}

func TestPublicStackOperationsRejectMissingIdentityBeforeCallingBackend(t *testing.T) {
	calls := withFakeRunner(t, func(string, ...string) cli.Output {
		t.Fatal("command runner called for an invalid stack")
		return cli.Output{}
	})

	for _, stack := range []*Stack{nil, &Stack{}} {
		if _, err := InstallHelmStack(stack); err == nil {
			t.Fatalf("install accepted invalid stack %#v", stack)
		}
		if _, err := UpgradeHelmStack(stack); err == nil {
			t.Fatalf("upgrade accepted invalid stack %#v", stack)
		}
		if err := DeleteHelmStack(stack); err == nil {
			t.Fatalf("delete accepted invalid stack %#v", stack)
		}
		if err := RollbackHelmStack(stack); err == nil {
			t.Fatalf("rollback accepted invalid stack %#v", stack)
		}
	}
	if len(*calls) != 0 {
		t.Fatalf("invalid stacks executed commands: %#v", *calls)
	}
}
