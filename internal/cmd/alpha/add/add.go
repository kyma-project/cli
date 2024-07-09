package add

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/add/managed"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewAddCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "add",
		Short:                 "Adds Kyma modules.",
		Long:                  `Use this command to add Kyma modules`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(managed.NewManagedCMD(kymaConfig))

	return cmd
}
