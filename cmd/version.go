package cmd

import (
	"fmt"

	"github.com/kyma-incubator/kymactl/internal"
	"github.com/spf13/cobra"
)

var Version string

var (
	kymaVersionCmd = []string{"get", "installation/kyma-installation", "-o", "jsonpath='{.spec.version}'"}
)

func newCmdVersion() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Version of the Kymactl and connected Kyma cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if Version == "" {
				fmt.Println("No Kymactl version available")
			} else {
				fmt.Printf("Kymactl version %s\n", Version)
			}

			kymaVersion := internal.RunKubeCmd(kymaVersionCmd)
			if kymaVersion == "" {
				fmt.Println("No Kyma cluster version available")
			} else {
				fmt.Printf("Kyma cluster version: %s\n", kymaVersion)
			}
		},
		Aliases: []string{"v"},
	}
	return versionCmd
}
