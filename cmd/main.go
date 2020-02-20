package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyma-project/cli/cmd/kyma"
	"github.com/kyma-project/cli/internal/cli"
)

func main() {
	SetupCloseHandler()
	command := kyma.NewCmd(cli.NewOptions())

	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal. Exiting...")
		os.Exit(0)
	}()
}
