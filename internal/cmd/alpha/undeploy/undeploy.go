package undeploy

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/undeploy/managed"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewUndeployCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "undeploy",
		Short:                 "Undeploy Kyma modules.",
		Long:                  `Use this command to undeploy Kyma modules`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(managed.NewManagedCMD(kymaConfig))

	return cmd
}
