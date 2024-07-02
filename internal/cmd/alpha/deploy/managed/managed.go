package managed

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kyma"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Short: "Deploy Kyma module on a managed Kyma instance",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runManagedDeploy(config)
		},
	}

	config.KubeClientConfig.AddFlag(cmd)
	cmd.Flags().StringVar(&config.module, "module", "", "Name of the module to deploy")
	cmd.Flags().StringVar(&config.channel, "channel", "", "Name of the Kyma channel to use for the module")

	_ = cmd.MarkFlagRequired("module")

	return cmd
}

func runManagedDeploy(config *managedConfig) error {
	kymaCR, err := kyma.GetDefaultKyma(config.Ctx, config.KubeClient)
	if err != nil {
		return err
	}

	spec := kymaCR.Object["spec"].(map[string]interface{})
	modules := make([]kyma.Module, 0)
	for _, m := range spec["modules"].([]interface{}) {
		modules = append(modules, kyma.ModuleFromInterface(m.(map[string]interface{})))
	}

	moduleExists := false
	for i, m := range modules {
		if m.Name == config.module {
			// module already exists, update channel
			modules[i].Channel = config.channel
			moduleExists = true
			break
		}
	}

	if !moduleExists {
		modules = append(modules, kyma.Module{
			Name:    config.module,
			Channel: config.channel,
		})
	}

	spec["modules"] = modules

	_, err = config.KubeClient.Dynamic().Resource(kyma.GVRKyma).
		Namespace("kyma-system").
		Update(config.Ctx, kymaCR, metav1.UpdateOptions{})
	return err
}
