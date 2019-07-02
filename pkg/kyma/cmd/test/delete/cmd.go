package del

import (
	"fmt"
	"time"

	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test/client"
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
		Use:     "delete",
		Short:   "Delete tests on a running Kyma cluster",
		Long:    `Delete tests on a running Kyma cluster`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"d"},
	}

	cobraCmd.Flags().StringVarP(&o.Name, "name", "n", "", "Test name to execute")
	cobraCmd.Flags().BoolVarP(&o.All, "all", "a", false, "Delete all test suites")
	return cobraCmd
}

func (cmd *command) Run() error {
	cli, err := client.NewTestRESTClient(10 * time.Second)
	if err != nil {
		return fmt.Errorf("unable to create test REST client. E: %s", err)
	}
	return cmd.deleteTestSuiteByName(cmd.opts.Name, cli)
}

func (cmd *command) deleteTestSuiteByName(name string, cli client.TestRESTClient) error {
	err := cli.DeleteTestSuite(test.NewTestSuite(name))
	if err != nil {
		return fmt.Errorf("unable to delete test suite '%s'. E: %s",
			cmd.opts.Name, err.Error())
	}
	fmt.Printf("test '%s' successfully delete\n", name)
	return nil
}
