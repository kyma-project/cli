package managed

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kyma"
	"github.com/spf13/cobra"
)

type managedConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	module  string
	channel string
}

func NewManagedCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := &managedConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "managed",
		Short: "Add managed Kyma module in a managed Kyma instance",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runAddManaged(config)
		},
	}

	config.KubeClientConfig.AddFlag(cmd)
	cmd.Flags().StringVar(&config.module, "module", "", "Name of the module to add")
	cmd.Flags().StringVar(&config.channel, "channel", "", "Name of the Kyma channel to use for the module")

	_ = cmd.MarkFlagRequired("module")

	return cmd
}

func runAddManaged(config *managedConfig) error {
	kymaCR, err := kyma.GetDefaultKyma(config.Ctx, config.KubeClient)
	if err != nil {
		return err
	}

	kymaCR = enableModule(kymaCR, config.module, config.channel)

	return kyma.UpdateDefaultKyma(config.Ctx, config.KubeClient, kymaCR)
}

func enableModule(kymaCR *kyma.Kyma, moduleName, moduleChannel string) *kyma.Kyma {
	for i, m := range kymaCR.Spec.Modules {
		if m.Name == moduleName {
			// module already exists, update channel
			kymaCR.Spec.Modules[i].Channel = moduleChannel
			return kymaCR
		}
	}

	kymaCR.Spec.Modules = append(kymaCR.Spec.Modules, kyma.Module{
		Name:    moduleName,
		Channel: moduleChannel,
	})

	return kymaCR
}
