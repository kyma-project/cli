package actions

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/actions/common"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
)

type resourceExplainActionConfig struct {
	Output string `yaml:"output"`
}

type resourceExplainAction struct {
	common.TemplateConfigurator[resourceExplainActionConfig]
}

func NewResourceExplain() types.Action {
	return &resourceExplainAction{}
}

func (a *resourceExplainAction) Run(cmd *cobra.Command, _ []string) clierror.Error {
	out.Msgln(a.Cfg.Output)
	return nil
}
