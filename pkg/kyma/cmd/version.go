package cmd

import (
	"fmt"

	"github.com/kyma-incubator/kyma-cli/internal"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"

	"github.com/spf13/cobra"
)

//Version contains the kyma-cli binary version injected by the build system
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
		Short: "Version of the kyma CLI and connected Kyma cluster",
		Long: `Prints the version of kyma CLI itself and the version of the kyma cluster connected by current KUBECONFIG
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
		version, err := internal.GetKymaVersion()
		if err != nil {
			return err
		}
		fmt.Printf("Kyma cluster version: %s\n", version)
	}

	return nil
}
