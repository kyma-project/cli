package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/cmd/kyma"
	"github.com/kyma-project/cli/internal/cli"
)

func main() {
	fmt.Print("WARNING: All commands within v2 are deprecated. We are designing v3 of kyma CLI  with a new set of commands that will be first released  within alpha command group.\n\n")

	command := kyma.NewCmd(cli.NewOptions())

	err := command.Execute()

	if err != nil {
		os.Exit(cli.GetExitCode(err))
	}
}
