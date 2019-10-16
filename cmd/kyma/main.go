package main

import (
	"os"

	"github.com/kyma-project/cli/internal/cli"
)

func main() {
	command := NewCmd(cli.NewOptions())

	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
