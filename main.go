package main

import (
	"github.com/kyma-project/cli.v3/internal/cmd"
)

func main() {
	cmd := cmd.NewKymaCMD()

	if err := cmd.Execute(); err != nil {
		panic(1)
	}
}
