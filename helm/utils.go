package helm

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const (
	maxHelmChartFiles      = 4096
	maxHelmChartFileBytes  = 8 << 20
	maxHelmChartTotalBytes = 64 << 20
	maxHelmChartPathBytes  = 4096
)

func validateStackIdentity(stack *Stack) error {
	if stack == nil {
		return fmt.Errorf("KubernetesStack cannot be nil")
	}
	if stack.Name == "" {
		return fmt.Errorf("KubernetesStack.Name cannot be empty")
	}
	return nil
}

func executeHelmCreateUpgradeTask(stack *Stack, args []string, isUpgrade bool) (string, error) {
	helmPath, cleanup, err := prepareHelmChart(stack)
	if err != nil {
		return "", err
	}
	defer cleanup()

	if stack.Namespace != "" {
		args = append(args, "--namespace", stack.Namespace)
	}
	if stack.Name != "" {
		if isUpgrade {
			args = append(args, stack.Name)
		} else { //create
			args = append(args, "--name", stack.Name)
		}
	} else {
		return "", fmt.Errorf("KubernetesStack.Name cannot be empty")
	}

	args = append(args, helmPath)
	if err := updateHelmDependencies(legacyHelm2Command, helmPath); err != nil {
		return "", err
	}
	output := runCommand(legacyHelm2Command, args...)
	if output.ExitCode > 0 {
		return "", fmt.Errorf("%s", output.StdErr)
	}
	return output.StdOut, output.Err
}

func prepareHelmChart(stack *Stack) (string, func(), error) {
	if err := validateStackIdentity(stack); err != nil {
		return "", nil, err
	}
	if len(stack.Files) > maxHelmChartFiles {
		return "", nil, fmt.Errorf("helm chart contains %d files; maximum is %d", len(stack.Files), maxHelmChartFiles)
	}

	names := make([]string, 0, len(stack.Files))
	totalBytes := 0
	for name, data := range stack.Files {
		if len(name) > maxHelmChartPathBytes {
			return "", nil, fmt.Errorf("helm chart file path exceeds %d bytes", maxHelmChartPathBytes)
		}
		if len(data) > maxHelmChartFileBytes {
			return "", nil, fmt.Errorf("helm chart file %q exceeds %d bytes", name, maxHelmChartFileBytes)
		}
		totalBytes += len(data)
		if totalBytes > maxHelmChartTotalBytes {
			return "", nil, fmt.Errorf("helm chart content exceeds %d bytes", maxHelmChartTotalBytes)
		}
		names = append(names, name)
	}
	sort.Strings(names)

	dir, err := os.MkdirTemp("", "helm-templates")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	err = os.Chmod(dir, 0700)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	chartDirectories := []string{}
	for _, name := range names {
		data := stack.Files[name]
		target, err := safeTemplatePath(dir, name)
		if err != nil {
			cleanup()
			return "", nil, err
		}

		cd := path.Dir(name)
		if cd == "." {
			cd = ""
		}
		err = os.MkdirAll(filepath.Dir(target), 0700)
		if err != nil {
			cleanup()
			return "", nil, err
		}
		if path.Base(name) == "Chart.yaml" {
			chartDirectories = append(chartDirectories, cd)
		}
		f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			cleanup()
			return "", nil, err
		}
		_, err = f.WriteString(data)
		closeErr := f.Close()
		if err != nil {
			cleanup()
			return "", nil, err
		}
		if closeErr != nil {
			cleanup()
			return "", nil, closeErr
		}
	}
	if len(chartDirectories) == 0 {
		cleanup()
		return "", nil, fmt.Errorf("helm chart is missing Chart.yaml")
	}
	sort.Slice(chartDirectories, func(i, j int) bool {
		leftDepth := chartDirectoryDepth(chartDirectories[i])
		rightDepth := chartDirectoryDepth(chartDirectories[j])
		if leftDepth == rightDepth {
			return chartDirectories[i] < chartDirectories[j]
		}
		return leftDepth < rightDepth
	})
	if len(chartDirectories) > 1 && chartDirectoryDepth(chartDirectories[0]) == chartDirectoryDepth(chartDirectories[1]) {
		cleanup()
		return "", nil, fmt.Errorf("helm chart contains multiple ambiguous root Chart.yaml files")
	}
	helmPath := path.Join(dir, chartDirectories[0])

	return helmPath, cleanup, nil
}

func chartDirectoryDepth(directory string) int {
	if directory == "" {
		return 0
	}
	return strings.Count(directory, "/") + 1
}

func updateHelmDependencies(command, helmPath string) error {
	output := runCommand(command, "dependency", "update", helmPath)
	if output.ExitCode > 0 {
		return fmt.Errorf("%s", output.StdErr)
	}
	if output.Err != nil {
		return output.Err
	}
	return nil
}

func safeTemplatePath(root, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("template file path cannot be empty")
	}
	if strings.Contains(name, "\\") {
		return "", fmt.Errorf("template file path %q contains a backslash", name)
	}

	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("template file path %q escapes the chart directory", name)
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	targetAbs := filepath.Join(rootAbs, clean)
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("template file path %q escapes the chart directory", name)
	}

	return targetAbs, nil
}

func collapseContiguousSpaces(s string) string {
	contiguousSpace := false
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			if contiguousSpace {
				return -1
			}
			contiguousSpace = true
			return ' '
		}
		contiguousSpace = false
		return r
	}, s)
}
