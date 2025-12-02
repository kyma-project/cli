package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewDashboardCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard <command> [flags]",
		Short: "Manages Kyma dashboard locally",
		Long:  `Use this command to manage running the Kyma dashboard locally in a docker container.`,
	}

	cmd.AddCommand(NewDashboardStartCMD(kymaConfig))
	cmd.AddCommand(NewDashboardStopCMD(kymaConfig))

	return cmd
}
