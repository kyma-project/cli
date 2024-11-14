package app

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewAppCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "app",
		Short:                 "Manage applications on the Kyma platform.",
		Long:                  `Use this command to manage applications on the Kyma platform.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewAppPushCMD(kymaConfig))

	return cmd
}
