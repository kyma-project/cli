package remove

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/remove/managed"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/communitymodules/cluster"
	"github.com/spf13/cobra"
)

type removeConfig struct {
	*cmdcommon.KymaConfig

	modules []string
}

func NewRemoveCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := removeConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove Kyma modules.",
		Long:  `Use this command to remove Kyma modules`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runRemove(&cfg))
		},
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(managed.NewManagedCMD(kymaConfig))
	cmd.Flags().StringSliceVar(&cfg.modules, "module", []string{}, "Name and version of the modules to remove. Example: --module serverless,keda:1.1.1,etc...")
	_ = cmd.MarkFlagRequired("module")

	return cmd
}

func runRemove(cfg *removeConfig) clierror.Error {
	modules := cluster.ParseModules(cfg.modules)
	client, err := cfg.GetKubeClientWithClierr()
	if err != nil {
		return err
	}

	return cluster.RemoveSpecifiedModules(cfg.Ctx, client.RootlessDynamic(), modules)
}
