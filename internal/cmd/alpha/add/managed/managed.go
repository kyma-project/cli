package managed

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kyma"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	kymaCR = updateCR(kymaCR, config.module, config.channel)

	_, err = kyma.UpdateDefaultKyma(config.Ctx, config.KubeClient, kymaCR)
	return err
}

func updateCR(kymaCR *unstructured.Unstructured, moduleName, moduleChannel string) *unstructured.Unstructured {
	newCR := kymaCR.DeepCopy()
	spec := newCR.Object["spec"].(map[string]interface{})
	modules := make([]kyma.Module, 0)
	for _, m := range spec["modules"].([]interface{}) {
		modules = append(modules, kyma.ModuleFromInterface(m.(map[string]interface{})))
	}

	moduleExists := false
	for i, m := range modules {
		if m.Name == moduleName {
			// module already exists, update channel
			modules[i].Channel = moduleChannel
			moduleExists = true
			break
		}
	}

	if !moduleExists {
		modules = append(modules, kyma.Module{
			Name:    moduleName,
			Channel: moduleChannel,
		})
	}

	spec["modules"] = modules

	return newCR
}
