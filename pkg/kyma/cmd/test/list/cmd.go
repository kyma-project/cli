package list

import (
	"fmt"

	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/spf13/cobra"
)

type command struct {
	opts *options
	core.Command
}

func NewCmd(o *options) *cobra.Command {
	cmd := command{
		Command: core.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "list",
		Short:   "Show available tests on a running Kyma cluster",
		Long:    `Show available tests on a running Kyma cluster`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"l"},
	}

	cobraCmd.Flags().BoolVarP(&o.Definitions, "definitions", "d", false, "Show test definitions only")
	cobraCmd.Flags().BoolVarP(&o.Tests, "tests", "t", false, "Show test-suites only")
	return cobraCmd
}

func (cmd *command) Run() error {
	if cmd.opts.Definitions {
		fmt.Println("Test definitions:")
		if testSuites, err := test.ListTestSuiteNames(cmd.Kubectl()); err != nil {
			return err
		} else {
			fmt.Println(testSuites)
		}
	}
	if cmd.opts.Tests {
		fmt.Println("Test-suites:")
		if testDefs, err := test.ListTestDefinitionNames(cmd.Kubectl()); err != nil {
			return err
		} else {
			fmt.Println(testDefs)
		}
	}
	return nil
}
