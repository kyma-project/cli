package main

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd"
)

func main() {
	cmd := cmd.NewKymaCMD()

	if err := cmd.Execute(); err != nil {
		clierror.Check(clierror.Wrap(err, clierror.New("failed to execute command", "use --help to see available commands and examples")))
	}
}
