package cmd

import (
	"github.com/kyma-project/cli.v3/internal/cmd/imageimport"
	"github.com/kyma-project/cli.v3/internal/cmd/provision"
	"github.com/spf13/cobra"
)

func NewKymaCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use: "kyma",

		// Affects children as well
		SilenceErrors: false,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, _ []string) {
			if err := cmd.Help(); err != nil {
				_ = err
			}
		},
	}

	cmd.AddCommand(provision.NewProvisionCMD())
	cmd.AddCommand(imageimport.NewImportCMD())

	return cmd
}
