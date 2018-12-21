package main

import (
	"fmt"
	"os"

	"github.com/kyma-incubator/kymactl/pkg/kyma/cmd"
)

func main() {
	command := cmd.NewKymaCmd(cmd.NewKymaOptions())

	err := command.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
