package helm

import (
	"testing"

	"github.com/PastureStack/kubectl-service/cli"
)

type commandCall struct {
	cmd  string
	args []string
}

func withFakeRunner(t *testing.T, fn func(cmd string, args ...string) cli.Output) *[]commandCall {
	t.Helper()

	old := runCommand
	calls := []commandCall{}
	runCommand = func(cmd string, args ...string) cli.Output {
		copiedArgs := append([]string(nil), args...)
		calls = append(calls, commandCall{cmd: cmd, args: copiedArgs})
		return fn(cmd, args...)
	}
	t.Cleanup(func() {
		runCommand = old
	})

	return &calls
}
