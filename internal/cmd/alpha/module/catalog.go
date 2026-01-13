package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/spf13/cobra"
)

type catalogConfig struct {
	*cmdcommon.KymaConfig

	remote       flags.BoolOrStrings
	outputFormat types.Format
}

func NewCatalogV2CMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := catalogConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "catalog [flags]",
		Short: "Lists modules catalog",
		Long:  `Use this command to list all available Kyma modules.`,
		Example: `  # List all modules available in the cluster (core and community)
  kyma module catalog

  # List available community modules from the official repository
  kyma module catalog --remote

  # List available community modules from a specific remote URL
  kyma module catalog --remote=https://example.com/modules.json

  # List available community modules from multiple remote URLs
  kyma module catalog --remote=https://example.com/modules1.json,https://example.com/modules2.json

  # Output catalog as JSON
  kyma module catalog -o json

  # List remote community modules in YAML format
  kyma module catalog --remote -o yaml`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(catalogModules(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (Possible values: table, json, yaml)")

	remoteFlag := cmd.Flags().VarPF(&cfg.remote, "remote", "", "Fetch modules from the official repository or specify custom URL(s)")
	remoteFlag.NoOptDefVal = "true"

	return cmd
}

func catalogModules(cfg *catalogConfig) clierror.Error {
	moduleOperations := modulesv2.NewModuleOperations(cmdcommon.NewKymaConfig())

	catalogOperation, err := moduleOperations.Catalog()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to execute the catalog command"))
	}

	catalogResult, err := catalogOperation.Run(cfg.Ctx, dtos.NewCatalogConfigFromRemote(cfg.remote))
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list available modules from the target Kyma environment"))
	}

	err = modulesv2.RenderCatalog(catalogResult, cfg.outputFormat)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to render catalog"))
	}

	return nil
}
