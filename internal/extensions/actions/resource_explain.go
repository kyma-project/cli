package actions

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
)

type resourceExplainActionConfig struct {
	Output string `yaml:"output"`
}

func NewResourceExplain(kymaConfig *cmdcommon.KymaConfig, actionConfig types.ActionConfig) types.CmdRun {
	return func(cmd *cobra.Command, args []string) {
		cfg := resourceExplainActionConfig{}
		clierror.Check(parseActionConfig(actionConfig, &cfg))
		fmt.Print(cfg.Output)
	}
}
