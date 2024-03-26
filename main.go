package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/cmd"
)

func main() {
	cmd := cmd.NewKymaCMD()

	if err := cmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
