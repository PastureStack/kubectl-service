package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestExecuteSuccessAndFailure(t *testing.T) {
	t.Setenv("GO_WANT_CLI_HELPER_PROCESS", "1")

	success := Execute(os.Args[0], "-test.run=TestCLIHelperProcess", "--", "success")
	if success.Err != nil || success.ExitCode != 0 {
		t.Fatalf("successful command returned err=%v exit=%d", success.Err, success.ExitCode)
	}
	if success.StdOut != "stdout-value" || success.StdErr != "stderr-value" {
		t.Fatalf("unexpected command streams: stdout=%q stderr=%q", success.StdOut, success.StdErr)
	}

	failure := Execute(os.Args[0], "-test.run=TestCLIHelperProcess", "--", "failure")
	if failure.Err == nil || failure.ExitCode != 7 {
		t.Fatalf("failed command returned err=%v exit=%d", failure.Err, failure.ExitCode)
	}
	if failure.StdOut != "partial-output" || failure.StdErr != "controlled-error" {
		t.Fatalf("unexpected failure streams: stdout=%q stderr=%q", failure.StdOut, failure.StdErr)
	}
}

func TestExecuteMissingCommand(t *testing.T) {
	output := Execute("pasturestack-command-that-does-not-exist")
	if output.Err == nil {
		t.Fatal("missing command unexpectedly succeeded")
	}
	if output.ExitCode != 0 {
		t.Fatalf("start failure must preserve the historical zero exit-code contract, got %d", output.ExitCode)
	}
	if output.StdOut != "" || output.StdErr != "" {
		t.Fatalf("missing command unexpectedly produced output: %#v", output)
	}
}

func TestErrExecUsesSanitizedStandardError(t *testing.T) {
	execErr := &ErrExec{Output: Output{StdErr: "catalog operation failed", Err: errors.New("internal")}}
	if got := execErr.Error(); got != "catalog operation failed" {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestCLIHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_HELPER_PROCESS") != "1" {
		return
	}
	separator := -1
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index
			break
		}
	}
	if separator < 0 || separator+1 >= len(os.Args) {
		fmt.Fprint(os.Stderr, "missing helper mode")
		os.Exit(9)
	}
	switch strings.TrimSpace(os.Args[separator+1]) {
	case "success":
		fmt.Fprint(os.Stdout, "stdout-value")
		fmt.Fprint(os.Stderr, "stderr-value")
		os.Exit(0)
	case "failure":
		fmt.Fprint(os.Stdout, "partial-output")
		fmt.Fprint(os.Stderr, "controlled-error")
		os.Exit(7)
	default:
		fmt.Fprint(os.Stderr, "unknown helper mode")
		os.Exit(8)
	}
}
