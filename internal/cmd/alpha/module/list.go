package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
)

func NewListV2CMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "Lists installed modules",
		Long: `Use this command to list all installed Kyma modules.

NOTE: functionality under construction
  - listing installed core modules: partial (name, version, channel)`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(listModulesV2(kymaConfig))
		},
	}

	return cmd
}

func listModulesV2(kymaConfig *cmdcommon.KymaConfig) clierror.Error {
	moduleOperations := modulesv2.NewModuleOperations(kymaConfig)

	listService, err := moduleOperations.List()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to execute the list command"))
	}

	results, err := listService.Run(kymaConfig.Ctx)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to list installed modules"))
	}

	err = modulesv2.RenderList(results, "", out.Default)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to render module list"))
	}

	return nil
}
