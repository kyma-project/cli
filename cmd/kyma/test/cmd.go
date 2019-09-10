package test

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests on a provisioned Kyma cluster.",
		Long: "Use this command to run tests on a provisioned Kyma cluster.",
	}
	cmd.Flags().Bool("help", false, "Displays help for the command.")
	return cmd
}
