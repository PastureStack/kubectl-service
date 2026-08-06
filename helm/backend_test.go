package helm

import (
	"strings"
	"testing"

	"github.com/PastureStack/kubectl-service/cli"
)

func TestActiveBackendIsExplicitLegacyHelm2(t *testing.T) {
	if got := ActiveBackendName(); got != LegacyHelm2BackendName {
		t.Fatalf("unexpected active backend %q", got)
	}
}

func TestConfigureBackendRequiresExplicitSupportedValue(t *testing.T) {
	if err := ConfigureBackend(Helm4BackendName); err != nil {
		t.Fatalf("configure Helm 4 backend: %v", err)
	}
	t.Cleanup(func() {
		if err := ConfigureBackend(LegacyHelm2BackendName); err != nil {
			t.Fatalf("restore legacy backend: %v", err)
		}
	})
	if got := ActiveBackendName(); got != Helm4BackendName {
		t.Fatalf("unexpected active backend %q", got)
	}

	err := ConfigureBackend("auto")
	if err == nil || !strings.Contains(err.Error(), "unsupported Helm backend") {
		t.Fatalf("unexpected invalid-backend result: %v", err)
	}
	if got := ActiveBackendName(); got != Helm4BackendName {
		t.Fatalf("invalid configuration changed active backend to %q", got)
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
