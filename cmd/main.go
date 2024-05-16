package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/cmd/kyma"
	"github.com/kyma-project/cli/internal/cli"
)

func main() {
	// TODO: improve warning message
	fmt.Println("WARNING: kyma command is deprecated and will be removed in the future")

	command := kyma.NewCmd(cli.NewOptions())

	err := command.Execute()

	if err != nil {
		os.Exit(cli.GetExitCode(err))
	}
}
