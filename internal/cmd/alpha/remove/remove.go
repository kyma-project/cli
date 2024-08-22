package remove

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/remove/managed"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewRemoveCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "remove",
		Short:                 "Remove Kyma modules.",
		Long:                  `Use this command to remove Kyma modules`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(managed.NewManagedCMD(kymaConfig))

	return cmd
}
