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
	setupCloseHandler()
	command := kyma.NewCmd(cli.NewOptions())

	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func setupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-c
		fmt.Printf("\r- Signal '%v' received from Terminal. Exiting...\n ", sig)
		os.Exit(0)
	}()
}
