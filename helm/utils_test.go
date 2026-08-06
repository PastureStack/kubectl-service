package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeTemplatePathAcceptsChartPaths(t *testing.T) {
	target, err := safeTemplatePath("/tmp/chart", "templates/deployment.yaml")
	if err != nil {
		t.Fatalf("expected chart path to be accepted: %v", err)
	}
	if target != "/tmp/chart/templates/deployment.yaml" {
		t.Fatalf("unexpected target path %q", target)
	}
}

func TestPrepareHelmChartChoosesUniqueShallowestChartAndCleansUp(t *testing.T) {
	stack := &Stack{
		Name: "demo",
		Files: map[string]string{
			"bundle/Chart.yaml":                "name: demo\nversion: 0.1.0\n",
			"bundle/charts/child/Chart.yaml":   "name: child\nversion: 0.1.0\n",
			"bundle/templates/deployment.yaml": "kind: Deployment\n",
		},
	}
	chartPath, cleanup, err := prepareHelmChart(stack)
	if err != nil {
		t.Fatalf("prepare chart: %v", err)
	}
	if filepath.Base(chartPath) != "bundle" {
		t.Fatalf("unexpected chart root %q", chartPath)
	}
	if _, err := os.Stat(chartPath); err != nil {
		t.Fatalf("chart root is not present: %v", err)
	}
	root := filepath.Dir(chartPath)
	cleanup()
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("chart temporary root was not removed: %v", err)
	}
}

func TestPrepareHelmChartRejectsAmbiguousRootsAndSizeLimits(t *testing.T) {
	_, _, err := prepareHelmChart(&Stack{
		Name: "demo",
		Files: map[string]string{
			"one/Chart.yaml": "name: one\nversion: 0.1.0\n",
			"two/Chart.yaml": "name: two\nversion: 0.1.0\n",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous chart roots returned %v", err)
	}

	_, _, err = prepareHelmChart(&Stack{
		Name: "demo",
		Files: map[string]string{
			"Chart.yaml": strings.Repeat("x", maxHelmChartFileBytes+1),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized chart file returned %v", err)
	}

	tooManyFiles := make(map[string]string, maxHelmChartFiles+1)
	for index := 0; index <= maxHelmChartFiles; index++ {
		tooManyFiles[fmt.Sprintf("templates/%04d.yaml", index)] = "kind: ConfigMap\n"
	}
	_, _, err = prepareHelmChart(&Stack{Name: "demo", Files: tooManyFiles})
	if err == nil || !strings.Contains(err.Error(), "maximum") {
		t.Fatalf("excessive chart file count returned %v", err)
	}

	_, _, err = prepareHelmChart(&Stack{
		Name: "demo",
		Files: map[string]string{
			strings.Repeat("x", maxHelmChartPathBytes+1): "data",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "path exceeds") {
		t.Fatalf("oversized chart path returned %v", err)
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
