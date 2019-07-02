package main

import (
	"os"

	"github.com/kyma-project/cli/pkg/kyma/cmd"
	"github.com/kyma-project/cli/pkg/kyma/core"
)

func main() {
	command := cmd.NewKymaCmd(core.NewOptions())

	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
