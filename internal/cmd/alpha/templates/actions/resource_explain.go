package actions

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type resourceExplainActionConfig struct {
	Output string `yaml:"output"`
}

func NewResourceExplain(kymaConfig *cmdcommon.KymaConfig, actionConfig types.ActionConfig) *cobra.Command {
	return &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			cfg := resourceExplainActionConfig{}
			clierror.Check(parseActionConfig(actionConfig, &cfg))
			fmt.Print(cfg.Output)
		},
	}
}
