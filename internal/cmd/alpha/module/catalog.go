package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/kyma-project/cli.v3/internal/output"
	"github.com/spf13/cobra"
)

type catalogConfig struct {
	*cmdcommon.KymaConfig
	outputFormat output.Format
}

func newCatalogCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := catalogConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "catalog [flags]",
		Short: "Lists modules catalog",
		Long:  `Use this command to list all available Kyma modules.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(catalogModules(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (Possible values: table, json, yaml)")

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

	err = modules.Render(modulesList, modules.CatalogTableInfo, cfg.outputFormat)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to render modules catalog"))
	}

	return nil
}
