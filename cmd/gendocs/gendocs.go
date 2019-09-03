package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/cmd/kyma"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra/doc"
)

const DocsTargetDir = "./docs"

func main() {
	command := kyma.NewCmd(cli.NewOptions())
	err := doc.GenMarkdownTree(command, DocsTargetDir)
	if err != nil {
		fmt.Println("unable to generate docs", err.Error())
		os.Exit(1)
	}
	fmt.Println("Docs successfully generated to the following dir", DocsTargetDir)
	os.Exit(0)
}
