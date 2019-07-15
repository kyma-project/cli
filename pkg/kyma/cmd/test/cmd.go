package test

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests on a provisioned Kyma cluster",
	}
	return cmd
}
