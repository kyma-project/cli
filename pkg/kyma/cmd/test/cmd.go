package test

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Operate with tests on provisioned Kyma cluster",
	}
	return cmd
}
