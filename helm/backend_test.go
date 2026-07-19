package helm

import "testing"

func TestActiveBackendIsExplicitLegacyHelm2(t *testing.T) {
	if got := ActiveBackendName(); got != LegacyHelm2BackendName {
		t.Fatalf("unexpected active backend %q", got)
	}
}
