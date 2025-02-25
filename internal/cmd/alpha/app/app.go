package app

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewAppCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "app <command> [flags]",
		Short:                 "Manages applications on the Kubernetes cluster",
		Long:                  `Use this command to manage applications on the Kubernetes cluster.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewAppPushCMD(kymaConfig))

	return cmd
}
