package del

import (
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
	return cobraCmd
}

func (cmd *command) Run() error {
	return nil
}
