package main

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd"
)

func main() {
	cmd, err := cmd.NewKymaCMD()
	clierror.Check(err)

	if err := cmd.Execute(); err != nil {
		clierror.Check(clierror.Wrap(err, clierror.New("failed to execute command")))
	}
}
