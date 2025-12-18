package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewDashboardStopCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [flags]",
		Short: `Terminates the locally running Kyma dashboard.`,
		Long:  `Use this command to terminate the locally running Kyma dashboard in a Docker container.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runDashboardStop())
		}}

	return cmd
}

func runDashboardStop() clierror.Error {
	//To be implemented in a future PR
	return nil
}
