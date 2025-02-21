package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
)

type deleteConfig struct {
	*cmdcommon.KymaConfig

	module string
}

func newDeleteCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := deleteConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:     "delete <module>",
		Short:   "Delete module",
		Aliases: []string{"del"},
		Long:    "Use this command to delete module",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runDelete(&cfg))
		},
	}

	return cmd
}

func runDelete(cfg *deleteConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	return modules.Disable(cfg.Ctx, client, cfg.module)
}
