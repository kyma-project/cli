package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/cmd/kyma"
	"github.com/kyma-project/cli/internal/cli"
)

func main() {
	fmt.Print("WARNING: All current commands are deprecated and new v3 commands will be developed within alpha command group.\n\n")

	command := kyma.NewCmd(cli.NewOptions())

	err := command.Execute()

	if err != nil {
		os.Exit(cli.GetExitCode(err))
	}
}
