package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
)

type modulesConfig struct {
	*cmdcommon.KymaConfig
}

func newListCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := modulesConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed modules",
		Long:  `List installed Kyma modules`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(listModules(&cfg))
		},
	}

	return cmd
}

func listModules(cfg *modulesConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	modulesList, err := modules.ListInstalled(cfg.Ctx, client)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list installed modules from the cluster"))
	}

	modules.Render(modulesList, modules.ModulesTableInfo)
	return nil
}
