package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
)

type removeConfig struct {
	*cmdcommon.KymaConfig

	module string
}

func newRemoveCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := removeConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "remove <module>",
		Short: "Remove module",
		Long:  "Use this command to remove module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runRemove(&cfg))
		},
	}

	return cmd
}

func runRemove(cfg *removeConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	return modules.Disable(cfg.Ctx, client, cfg.module)
}
