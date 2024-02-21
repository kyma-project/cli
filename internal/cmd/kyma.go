package cmd

import (
	"github.com/kyma-project/cli.v3/internal/cmd/provision"
	"github.com/spf13/cobra"
)

func NewKymaCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use: "kyma",
		// 		Short: "Controls a Kyma cluster.",
		// 		Long: `Kyma is a flexible and easy way to connect and extend enterprise applications in a cloud-native world.
		// Kyma CLI allows you to install and manage Kyma.

		// `,

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

	return cmd
}
