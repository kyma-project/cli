package cmd

import (
	"fmt"

	"github.com/kyma-incubator/kymactl/internal"

	"github.com/spf13/cobra"
)

//Version contains the kymactl binary version injected by the build system
var Version string

//VersionOptions defines available options for the command
type VersionOptions struct {
	Client bool
}

//NewVersionOptions creates options with default values
func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

//NewVersionCmd creates a new version command
func NewVersionCmd(o *VersionOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Version of the kymactl and connected Kyma cluster",
		Long: `Prints the version of kymactl itself and the version of the kyma cluster connected by current KUBECONFIG
`,
		RunE: func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	cmd.Flags().BoolVarP(&o.Client, "client", "c", false, "Client version only (no server required)")

	return cmd
}

//Run runs the command
func (o *VersionOptions) Run() error {

	version := Version
	if version == "" {
		version = "N/A"
	}
	fmt.Printf("Kymactl version:      %s\n", version)

	if !o.Client {
		version, err := internal.GetKymaVersion()
		if err != nil {
			return err
		}
		fmt.Printf("Kyma cluster version: %s\n", version)
	}

	return nil
}
