package helm

import "testing"

func TestSafeTemplatePathAcceptsChartPaths(t *testing.T) {
	target, err := safeTemplatePath("/tmp/chart", "templates/deployment.yaml")
	if err != nil {
		t.Fatalf("expected chart path to be accepted: %v", err)
	}
	if target != "/tmp/chart/templates/deployment.yaml" {
		t.Fatalf("unexpected target path %q", target)
	}
}

func TestSafeTemplatePathRejectsTraversal(t *testing.T) {
	cases := []string{
		"",
		"../Chart.yaml",
		"templates/../../secret",
		"/etc/passwd",
		`templates\secret.yaml`,
	}

	for _, tc := range cases {
		if _, err := safeTemplatePath("/tmp/chart", tc); err == nil {
			t.Fatalf("expected %q to be rejected", tc)
		}
	}
}
