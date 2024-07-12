package managed

import (
	"slices"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/spf13/cobra"
)

type managedConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	module string
}

func NewManagedCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := &managedConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "managed",
		Short: "Remove Kyma module on a managed Kyma instance",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runRemoveManaged(config)
		},
	}

	config.KubeClientConfig.AddFlag(cmd)
	cmd.Flags().StringVar(&config.module, "module", "", "Name of the module to remove")

	_ = cmd.MarkFlagRequired("module")

	return cmd
}

func runRemoveManaged(config *managedConfig) error {
	kymaCR, err := config.KubeClient.Kyma().GetDefaultKyma(config.Ctx)
	if err != nil {
		return err
	}

	kymaCR = disableModule(kymaCR, config.module)

	return config.KubeClient.Kyma().UpdateDefaultKyma(config.Ctx, kymaCR)
}

func disableModule(kymaCR *kyma.Kyma, moduleName string) *kyma.Kyma {
	kymaCR.Spec.Modules = slices.DeleteFunc(kymaCR.Spec.Modules, func(m kyma.Module) bool {
		return m.Name == moduleName
	})

	return kymaCR
}
