package main

import (
	"os"

	"github.com/kyma-project/cli/cmd/kyma"
	"github.com/kyma-project/cli/internal/cli"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	command := kyma.NewCmd(cli.NewOptions())

	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
