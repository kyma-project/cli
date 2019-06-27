package cmd

import (
	"fmt"

	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/pkg/kyma/core"

	"github.com/spf13/cobra"
)

//Version contains the cli binary version injected by the build system
var Version string

//VersionOptions defines available options for the command
type VersionOptions struct {
	*core.Options
	Client bool
}

//NewVersionOptions creates options with default values
func NewVersionOptions(o *core.Options) *VersionOptions {
	return &VersionOptions{Options: o}
}

//NewVersionCmd creates a new version command
func NewVersionCmd(o *VersionOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Version of the Kyma CLI and connected Kyma cluster",
		Long: `Prints the version of the Kyma CLI and the version of the Kyma cluster the current KUBECONFIG points to.

		Prerequisites
		- Kubernetes cluster is available with the KUBECONFIG already pointing to it (optional).

		Use cases
		The following examples include the most common cases of using the version command.
		1. Print out client and server version:
		kyma version

		2. Print out only the client version:
		kyma version --client
	
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
	fmt.Printf("Kyma CLI version: %s\n", version)

	if !o.Client {
		version, err := internal.GetKymaVersion(o.Verbose)
		if err != nil {
			fmt.Printf("Unable to get Kyma cluster version due to error: %s. Please check if your cluster is available and has Kyma installed\r\n", err.Error())
			return nil
		}
		fmt.Printf("Kyma cluster version: %s\n", version)
	}

	return nil
}
