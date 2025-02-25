package module

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewModuleCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "module <command> [flags]",
		Aliases: []string{"modules"},
		Short:   "Manages kyma modules",
		Long:    `Use this command to manage modules in the Kyma cluster.`,
	}

	cmd.AddCommand(newListCMD(kymaConfig))
	cmd.AddCommand(newCatalogCMD(kymaConfig))
	cmd.AddCommand(newAddCMD(kymaConfig))
	cmd.AddCommand(newDeleteCMD(kymaConfig))
	cmd.AddCommand(newManageCMD(kymaConfig))
	cmd.AddCommand(newUnmanageCMD(kymaConfig))

	return cmd
}
