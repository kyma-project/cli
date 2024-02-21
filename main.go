package main

import (
	"os"

	"github.com/kyma-project/cli.v3/internal/cmd"
)

func main() {
	cmd := cmd.NewKymaCMD()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
