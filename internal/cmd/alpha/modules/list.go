package modules

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
)

type modulesConfig struct {
	*cmdcommon.KymaConfig
}

func NewListCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := modulesConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List modules.",
		Long:  `List either installed, managed or available Kyma modules.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(listModules(&cfg))
		},
	}

	return cmd
}

// listModules collects all the methods responsible for the command and its flags
func listModules(cfg *modulesConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	modulesList, err := modules.List(cfg.Ctx, client)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list available modules from the cluster"))
	}

	modules.Render(modulesList, modules.ModulesTableInfo)
	return nil
}
