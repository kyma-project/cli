package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd"
)

func main() {
	cmd, err := cmd.NewKymaCMD()
	clierror.Check(err)

	if err := cmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
