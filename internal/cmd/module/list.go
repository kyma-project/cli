package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/kyma-project/cli.v3/internal/modulesv2/precheck"
	"github.com/spf13/cobra"
)

type modulesConfig struct {
	*cmdcommon.KymaConfig
	outputFormat types.Format
	showErrors   bool
}

func newListCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := modulesConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "Lists the installed modules",
		Long:  `Use this command to list the installed Kyma modules.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(precheck.RequireCRD(kymaConfig, precheck.CmdGroupStable))
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(listModules(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (Possible values: table, json, yaml)")
	cmd.Flags().BoolVar(&cfg.showErrors, "show-errors", false, "Indicates whether to show errors outputted by misconfigured modules")

	return cmd
}

func listModules(cfg *modulesConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}
	moduleTemplatesRepo := repo.NewModuleTemplatesRepo(client)

	modulesList, err := modules.ListInstalled(cfg.Ctx, client, moduleTemplatesRepo, cfg.showErrors)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list installed modules from the target Kyma environment"))
	}

	err = modules.Render(modulesList, modules.ModulesTableInfo, cfg.outputFormat)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to render module catalog"))
	}

	return nil
}
