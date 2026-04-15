package module

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/spf13/cobra"
)

func NewListV2CMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "Lists installed modules",
		Long: `Use this command to list all installed Kyma modules.

NOTE: functionality under construction
  - listing installed modules: partial (names only)`,
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

	for _, r := range results {
		fmt.Println(r.Name)
	}

	return nil
}
