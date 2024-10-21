package managed

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type managedConfig struct {
	*cmdcommon.KymaConfig

	module  string
	channel string
}

func NewManagedCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := &managedConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "managed",
		Short: "Add managed Kyma module in a managed Kyma instance",
		PreRun: func(_ *cobra.Command, _ []string) {
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runAddManaged(config)
		},
	}

	cmd.Flags().StringVar(&config.module, "module", "", "Name of the module to add")
	cmd.Flags().StringVar(&config.channel, "channel", "", "Name of the Kyma channel to use for the module")

	_ = cmd.MarkFlagRequired("module")

	return cmd
}

func runAddManaged(config *managedConfig) error {
	return config.KubeClient.Kyma().EnableModule(config.Ctx, config.module, config.channel)
}
