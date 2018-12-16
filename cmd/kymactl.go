package main

import (
	"fmt"
	"os"

	"github.com/kyma-incubator/kymactl/pkg/kymactl/cmd"
)

func main() {
	command := cmd.NewKymactlCmd(cmd.NewKymactlOptions())

	err := command.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
