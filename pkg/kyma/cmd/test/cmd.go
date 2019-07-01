package test

import (
	"github.com/spf13/cobra"
)

const TestCrdDefinition = "testdefinitions.testing.kyma-project.io"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Operate with tests on provisioned Kyma cluster",
	}
	return cmd
}
