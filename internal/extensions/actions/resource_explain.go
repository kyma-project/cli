package actions

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
)

type resourceExplainActionConfig struct {
	Output string `yaml:"output"`
}

type resourceExplainAction struct {
	configurator[resourceExplainActionConfig]
}

func NewResourceExplain() types.Action {
	return &resourceExplainAction{}
}

func (a *resourceExplainAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	fmt.Fprint(cmd.OutOrStdout(), a.cfg.Output)
	return nil
}
