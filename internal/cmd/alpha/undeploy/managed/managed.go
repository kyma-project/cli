package managed

import (
	"slices"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kyma"
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
		Short: "Undeploy Kyma module on a managed Kyma instance",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runManagedUndeploy(config)
		},
	}

	config.KubeClientConfig.AddFlag(cmd)
	cmd.Flags().StringVar(&config.module, "module", "", "Name of the module to undeploy")

	_ = cmd.MarkFlagRequired("module")

	return cmd
}

func runManagedUndeploy(config *managedConfig) error {
	kymaCR, err := kyma.GetDefaultKyma(config.Ctx, config.KubeClient)
	if err != nil {
		return err
	}

	spec := kymaCR.Object["spec"].(map[string]interface{})
	modules := make([]kyma.Module, 0)
	for _, m := range spec["modules"].([]interface{}) {
		modules = append(modules, kyma.ModuleFromInterface(m.(map[string]interface{})))
	}

	modules = slices.DeleteFunc(modules, func(m kyma.Module) bool {
		return m.Name == config.module
	})

	spec["modules"] = modules

	_, err = kyma.UpdateDefaultKyma(config.Ctx, config.KubeClient, kymaCR)
	return err
}
