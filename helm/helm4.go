package helm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PastureStack/kubectl-service/cli"
)

const helm4ListLimit = 10000

type helm4ListItem struct {
	Name       string          `json:"name"`
	Namespace  string          `json:"namespace"`
	Revision   json.RawMessage `json:"revision"`
	Updated    string          `json:"updated"`
	Status     string          `json:"status"`
	Chart      string          `json:"chart"`
	AppVersion string          `json:"app_version"`
}

func executeHelm4CreateUpgradeTask(stack *Stack, isUpgrade bool) (string, error) {
	helmPath, cleanup, err := prepareHelmChart(stack)
	if err != nil {
		return "", err
	}
	defer cleanup()

	if err := updateHelmDependencies(helm4Command, helmPath); err != nil {
		return "", err
	}

	action := "install"
	if isUpgrade {
		action = "upgrade"
	}
	args := []string{action, stack.Name, helmPath}
	if stack.Namespace != "" {
		args = append(args, "--namespace", stack.Namespace)
	}
	// Helm 4 defaults new releases to server-side apply. The compatibility
	// service preserves the client-side behavior used by existing charts.
	args = append(args, "--server-side=false", "--color=never")

	output := runCommand(helm4Command, args...)
	if output.ExitCode > 0 {
		return "", fmt.Errorf("%s", output.StdErr)
	}
	return output.StdOut, output.Err
}

func listReleasesHelm4() ([]Release, error) {
	output := runCommand(
		helm4Command,
		"list", "--all-namespaces", "--output", "json",
		"--max", strconv.Itoa(helm4ListLimit), "--color=never",
	)
	if output.ExitCode > 0 {
		return nil, &cli.ErrExec{Output: output}
	}
	if output.Err != nil {
		return nil, output.Err
	}
	if strings.TrimSpace(output.StdOut) == "" {
		return []Release{}, nil
	}

	items := []helm4ListItem{}
	if err := json.Unmarshal([]byte(output.StdOut), &items); err != nil {
		return nil, fmt.Errorf("parse Helm 4 release list: %w", err)
	}
	if len(items) >= helm4ListLimit {
		return nil, fmt.Errorf("Helm 4 release list reached the safety limit of %d; narrow the migration scope", helm4ListLimit)
	}

	releases := make([]Release, 0, len(items))
	for _, item := range items {
		if item.Name == "" {
			return nil, fmt.Errorf("parse Helm 4 release list: release name is empty")
		}
		if item.Namespace == "" {
			return nil, fmt.Errorf("parse Helm 4 release %q: namespace is empty", item.Name)
		}
		revision, err := parseHelm4Revision(item.Revision)
		if err != nil {
			return nil, fmt.Errorf("parse Helm 4 release %q revision: %w", item.Name, err)
		}
		updated, err := parseHelm4Time(item.Updated)
		if err != nil {
			return nil, fmt.Errorf("parse Helm 4 release %q timestamp: %w", item.Name, err)
		}
		releases = append(releases, Release{
			Name:      item.Name,
			Namespace: item.Namespace,
			Revision:  revision,
			Updated:   updated,
			Status:    strings.ToUpper(item.Status),
			Chart:     item.Chart,
		})
	}

	return releases, nil
}

func parseHelm4Revision(raw json.RawMessage) (int, error) {
	value := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	if value == "" || value == "null" {
		return 0, fmt.Errorf("revision is empty")
	}
	revision, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if revision < 1 {
		return 0, fmt.Errorf("revision must be positive")
	}
	return revision, nil
}

func parseHelm4Time(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp %q", value)
}

func findHelm4Release(stack *Stack) (*Release, error) {
	releases, err := listReleasesHelm4()
	if err != nil {
		return nil, err
	}
	matches := []Release{}
	for _, release := range releases {
		if release.Name != stack.Name {
			continue
		}
		if stack.Namespace != "" && release.Namespace != stack.Namespace {
			continue
		}
		matches = append(matches, release)
	}
	if len(matches) == 0 {
		return nil, nil
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("Helm 4 release %q is ambiguous across namespaces; provide KubernetesStack.Namespace", stack.Name)
	}
	return &matches[0], nil
}

func deleteHelmStackHelm4(stack *Stack) error {
	release, err := findHelm4Release(stack)
	if err != nil || release == nil {
		return err
	}

	args := []string{"uninstall", stack.Name, "--namespace", release.Namespace, "--color=never"}
	output := runCommand(helm4Command, args...)
	if output.ExitCode > 0 {
		return fmt.Errorf("%s", output.StdErr)
	}
	if output.Err != nil {
		return output.Err
	}
	if stack.Namespace != "" {
		return deleteKubernetesNamespace(stack.Namespace)
	}
	return nil
}

func rollbackHelmStackHelm4(stack *Stack) error {
	args := []string{"rollback", stack.Name, "0"}
	if stack.Namespace != "" {
		args = append(args, "--namespace", stack.Namespace)
	}
	args = append(args, "--server-side=false", "--color=never")
	output := runCommand(helm4Command, args...)
	if output.ExitCode > 0 {
		return fmt.Errorf("%s", output.StdErr)
	}
	return output.Err
}
