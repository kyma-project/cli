package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
)

type listConfig struct {
	*cmdcommon.KymaConfig
	outputFormat types.Format
}

func NewListV2CMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := listConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "Lists installed modules",
		Long: `Use this command to list the installed Kyma modules.

NOTE: functionality under construction
  - community modules not yet supported`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(listModulesV2(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (Possible values: table, json, yaml)")

	return cmd
}

func listModulesV2(cfg *listConfig) clierror.Error {
	moduleOperations := modulesv2.NewModuleOperations(cfg.KymaConfig)

	listService, err := moduleOperations.List()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to execute the list command"))
	}

	results, err := listService.Run(cfg.Ctx)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list installed modules"))
	}

	err = modulesv2.RenderList(results, cfg.outputFormat, out.Default)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to render module list"))
	}

	return nil
}
