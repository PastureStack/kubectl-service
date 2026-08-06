package helm

import "github.com/PastureStack/kubectl-service/cli"

type commandRunner func(cmd string, args ...string) cli.Output

var runCommand commandRunner = cli.Execute
