package events

import (
	"strings"
	"testing"
)

func TestSafeDataKeysDoesNotExposeValues(t *testing.T) {
	got := safeDataKeys(map[string]interface{}{
		"secret": "super-secret-value",
		"name":   "catalog",
	})

	if strings.Contains(got, "super-secret-value") {
		t.Fatalf("safe data keys exposed a secret value: %q", got)
	}
	if got != "name,secret" {
		t.Fatalf("unexpected keys: %q", got)
	}
}
