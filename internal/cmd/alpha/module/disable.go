package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
)

type disableConfig struct {
	*cmdcommon.KymaConfig

	module string
}

func newDisableCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := disableConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "disable <module>",
		Short: "Disable module",
		Long:  "Use this command to disable module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runDisable(&cfg))
		},
	}

	return cmd
}

func runDisable(cfg *disableConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	return modules.Disable(cfg.Ctx, client, cfg.module)
}
