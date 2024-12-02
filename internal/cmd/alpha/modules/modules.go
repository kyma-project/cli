package modules

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewModulesCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "module",
		Aliases: []string{"modules"},
		Short:   "Manage kyma modules.",
		Long:    `Use this command to manage modules on a kyma cluster.`,
	}

	cmd.AddCommand(NewListCMD(kymaConfig))

	return cmd
}
