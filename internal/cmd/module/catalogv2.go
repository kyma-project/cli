package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/di"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/spf13/cobra"
)

type catalogV2Config struct {
	*cmdcommon.KymaConfig
	outputFormat types.Format
}

func newCatalogV2CMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := catalogV2Config{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "catalogv2 [flags]",
		Short: "Lists modules catalog",
		Long:  `Use this command to list all available Kyma modules.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(catalogV2Modules(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (Possible values: table, json, yaml)")

	return cmd
}

func catalogV2Modules(cfg *catalogV2Config) clierror.Error {
	c, err := modulesv2.SetupDIContainer(cfg.KymaConfig)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to configure command dependencies"))
	}

	catalogService, err := di.GetTyped[*modulesv2.CatalogService](c)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to execute the catalog command"))
	}

	catalogResult, err := catalogService.Run(cfg.Ctx, []string{"https://kyma-project.github.io/community-modules/all-modules.json"})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list available modules from the target Kyma environment"))
	}

	err = modulesv2.RenderCatalog(catalogResult, cfg.outputFormat)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to render catalog"))
	}

	return nil
}
