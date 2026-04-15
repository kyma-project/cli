package module

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
)

func NewListV2CMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "Lists installed modules",
		Long:  `Use this command to list all installed Kyma modules.`,
		Run: func(_ *cobra.Command, _ []string) {
			out.Default.Msgln("functionality under construction")
		},
	}

	return cmd
}
