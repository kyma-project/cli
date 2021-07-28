package test

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Runs tests on a provisioned Kyma cluster.",
		Long:  "Use this command to run tests on a provisioned Kyma cluster.",
		Deprecated: "test is deprecated!",
	}
	return cmd
}
