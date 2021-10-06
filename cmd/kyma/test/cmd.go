package test

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "[DEPRECATED] Runs tests on a provisioned Kyma cluster.",
		Long: `[DEPRECATED: The "test" command works only with Kyma 1.x.x.]

Use this command to run tests on a provisioned Kyma cluster.`,
	}
	return cmd
}
