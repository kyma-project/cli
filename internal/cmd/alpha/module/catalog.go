package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/spf13/cobra"
)

type catalogV2Config struct {
	*cmdcommon.KymaConfig

	origin       []string
	outputFormat types.Format
}

func NewCatalogV2CMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := catalogV2Config{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "catalog [flags]",
		Short: "Lists modules catalog",
		Long:  `Use this command to list all available Kyma modules.`,
		Example: `
  # List all available modules from all origins
  kyma module catalog

  # List only official Kyma modules managed by KLM with SLA
  kyma module catalog --origin kyma

  # List only community modules (not officially supported)
  kyma module catalog --origin community

  # List only community modules already available on the cluster
  kyma module catalog --origin cluster

  # List modules from multiple origins
  kyma module catalog --origin kyma,community

  # Output catalog as JSON
  kyma module catalog -o json

  # List official Kyma modules in YAML format
  kyma module catalog --origin kyma -o yaml`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(catalogModules(&cfg))
		},
	}

	cmd.Flags().StringSliceVar(&cfg.origin, "origin", []string{"kyma", "community", "cluster"}, "Specifies the source of the module")
	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (Possible values: table, json, yaml)")

	return cmd
}

func catalogModules(cfg *catalogV2Config) clierror.Error {
	moduleOperations := modulesv2.NewModuleOperations(cmdcommon.NewKymaConfig())

	catalogOperation, err := moduleOperations.Catalog()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to execute the catalog command"))
	}

	catalogResult, err := catalogOperation.Run(cfg.Ctx, dtos.NewCatalogConfigFromOriginsList(cfg.origin))
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list available modules from the target Kyma environment"))
	}

	err = modulesv2.RenderCatalog(catalogResult, cfg.outputFormat)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to render catalog"))
	}

	return nil
}
