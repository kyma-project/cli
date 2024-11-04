package managed

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type managedConfig struct {
	*cmdcommon.KymaConfig

	module string
}

func NewManagedCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := &managedConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "managed",
		Short: "Remove Kyma module on a managed Kyma instance",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runRemoveManaged(config)
		},
	}

	cmd.Flags().StringVar(&config.module, "module", "", "Name of the module to remove")

	_ = cmd.MarkFlagRequired("module")

	return cmd
}

func runRemoveManaged(config *managedConfig) error {
	client, err := config.GetKubeClient()
	if err != nil {
		return err
	}

	return client.Kyma().DisableModule(config.Ctx, config.module)
}
