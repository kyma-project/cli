package module

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewModuleCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "module <command> [flags]",
		Aliases: []string{"modules"},
		Short:   "Manages Kyma modules",
		Long:    `Use this command to manage modules in the Kyma cluster.`,
	}

	cmd.AddCommand(NewCatalogV2CMD(kymaConfig))
	cmd.AddCommand(NewPullV2CMD(kymaConfig))

	return cmd
}
