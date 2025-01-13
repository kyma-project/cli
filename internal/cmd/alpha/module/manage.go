package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type manageConfig struct {
	*cmdcommon.KymaConfig

	module string
}

func newManageCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := manageConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "manage <module>",
		Short: "Manage module.",
		Long:  "Use this command to manage an existing module.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runManage(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.module, "module", "", "Name of the module to manage")
	cmd.MarkFlagRequired("module")
	return cmd
}

func runManage(cfg *manageConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}
	err := client.Kyma().ManageModule(cfg.Ctx, cfg.module, true)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to set module as managed"))
	}
	client.Kyma().WaitForModuleState()
	return nil
}
