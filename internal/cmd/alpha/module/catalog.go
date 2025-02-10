package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
)

type catalogConfig struct {
	*cmdcommon.KymaConfig
}

func newCatalogCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := catalogConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "List modules catalog.",
		Long:  `List available Kyma modules.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(catalogModules(&cfg))
		},
	}

	return cmd
}

func catalogModules(cfg *catalogConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	modulesList, err := modules.ListCatalog(cfg.Ctx, client)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list available modules from the cluster"))
	}

	modules.Render(modulesList, modules.CatalogTableInfo)
	return nil
}
