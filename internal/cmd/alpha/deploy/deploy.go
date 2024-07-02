package deploy

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/deploy/managed"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewDeployCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "deploy",
		Short:                 "Deploy Kyma modules.",
		Long:                  `Use this command to deploy Kyma modules`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(managed.NewManagedCMD(kymaConfig))

	return cmd
}
