package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/kyma-project/cli.v3/internal/output"
	"github.com/spf13/cobra"
)

type modulesConfig struct {
	*cmdcommon.KymaConfig
	outputFormat output.Format
}

func newListCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := modulesConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "Lists the installed modules",
		Long:  `Use this command to list the installed Kyma modules.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(listModules(&cfg))
		},
	}

	cmd.Flags().VarP(&cfg.outputFormat, "output", "o", "Output format (Possible values: table, json, yaml)")

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

	err = modules.Render(modulesList, modules.ModulesTableInfo, cfg.outputFormat)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to render modules catalog"))
	}

	return nil
}
