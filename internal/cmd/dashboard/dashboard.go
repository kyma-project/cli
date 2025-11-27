package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewDashboardCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Runs the Kyma dashboard locally and opens it directly in a web browser.",
		Long:  `Use this command to run the Kyma dashboard locally in a docker container and open it directly in a web browser. This command only works with a local installation of Kyma.`,
	}

	cmd.AddCommand(NewDashboardStartCMD(kymaConfig))

	return cmd
}
