package test

import (
	"github.com/spf13/cobra"
)

const TestCrdDefinition = "testdefinitions.testing.kyma-project.io"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests on a provisioned Kyma cluster",
	}
	return cmd
}
