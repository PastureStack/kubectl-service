package helm

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

func executeHelmCreateUpgradeTask(stack *Stack, args []string, isUpgrade bool) (string, error) {
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
	dir, err := os.MkdirTemp("", "helm-templates")
	if err != nil {
		return "", err
	}
	err = os.Chmod(dir, 0700)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	helmPath := ""
	for name, data := range stack.Files {
		target, err := safeTemplatePath(dir, name)
		if err != nil {
			return "", err
		}

		cd := path.Dir(name)
		if cd == "." {
			cd = ""
		}
		err = os.MkdirAll(filepath.Dir(target), 0700)
		if err != nil {
			return "", err
		}
		if strings.HasSuffix(name, "Chart.yaml") {
			helmPath = path.Join(dir, cd)
		}
		f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return "", err
		}
		_, err = f.WriteString(data)
		closeErr := f.Close()
		if err != nil {
			return "", err
		}
		if closeErr != nil {
			return "", closeErr
		}
	}
	if helmPath == "" {
		return "", fmt.Errorf("helm chart is missing Chart.yaml")
	}
	args = append(args, helmPath)
	depArgs := []string{"dependency", "update", helmPath}
	output := runCommand(legacyHelm2Command, depArgs...)
	if output.ExitCode > 0 {
		return "", fmt.Errorf("%s", output.StdErr)
	}
	if output.Err != nil {
		return "", output.Err
	}
	output = runCommand(legacyHelm2Command, args...)
	if output.ExitCode > 0 {
		return "", fmt.Errorf("%s", output.StdErr)
	}
	return output.StdOut, output.Err
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
