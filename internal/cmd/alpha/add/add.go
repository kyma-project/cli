package add

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/remove/managed"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/communitymodules/cluster"
	"github.com/spf13/cobra"
)

type addConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	modules []string
	crs     []string
}

func NewAddCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := addConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adds Kyma modules.",
		Long:  `Use this command to add Kyma modules`,
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(cfg.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runAdd(&cfg))
		},
	}

	cmd.AddCommand(managed.NewManagedCMD(kymaConfig))

	cfg.KubeClientConfig.AddFlag(cmd)
	cmd.Flags().StringSliceVar(&cfg.modules, "module", []string{}, "Name and version of the modules to add. Example: --module serverless,keda:1.1.1,etc...")
	cmd.Flags().StringSliceVar(&cfg.crs, "cr", []string{}, "Path to the custom CR file")

	return cmd
}

func runAdd(cfg *addConfig) clierror.Error {
	err := cluster.AssureNamespace(cfg.Ctx, cfg.KubeClient.Static(), "kyma-system")
	if err != nil {
		return err
	}

	return cluster.ApplySpecifiedModules(cfg.Ctx, cfg.KubeClient.RootlessDynamic(), cfg.modules, cfg.crs)
}
