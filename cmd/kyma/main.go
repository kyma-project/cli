package main

import (
	"os"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/cmd"
)

func main() {
	command := cmd.NewCmd(cli.NewOptions())

	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
