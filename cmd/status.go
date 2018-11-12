package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/clebs/ldng"
	"github.com/kyma-incubator/kymactl/internal"
	"github.com/spf13/cobra"
)

const (
	sleep = 10 * time.Second
)

var (
	statusCmd = []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'"}
	descCmd   = []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'"}
)

func newCmdStatus() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check the status of Kyma",
		Run:   checkStatus,
	}
	return statusCmd
}

func checkStatus(cmd *cobra.Command, args []string) {
	currentDesc := ""
	var stop chan struct{}

	for {
		status := internal.RunKubeCmd(statusCmd)
		desc := internal.RunKubeCmd(descCmd)

		switch status {
		case "Installed":
			if stop != nil {
				stop <- struct{}{} // stop the spinner
				<-stop             // wait for the spinner to finish cleanup
				stop = nil
			}
			fmt.Println("Kyma is running!")
			// TODO maybe run kubectl cluster-info here
			os.Exit(0)

		case "Error":
			fmt.Printf("Error installing Kyma: %s\n", desc)
			// TODO print logs here
			os.Exit(1)

		case "InProgress":
			// only do something if the description has changed
			if desc != currentDesc {
				if stop != nil {
					stop <- struct{}{} // stop the spinner
					<-stop             // wait for the spinner to finish cleanup
					stop = nil
					continue
				} else {
					s := ldng.NewSpin(ldng.SpinPrefix(fmt.Sprintf("%s", desc)), ldng.SpinPeriod(100*time.Millisecond), ldng.SpinSuccess(fmt.Sprintf("%s ✅\n", desc)))
					stop = s.Start()
					currentDesc = desc
				}
			}

		default:
			fmt.Printf("Unexpected status: %s\n", status)
			break
		}
		time.Sleep(sleep)
	}
}
