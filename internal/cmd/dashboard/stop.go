package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewDashboardStopCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [flags]",
		Short: "Run Kyma dashboard locally",
		Long:  `Use this command to run the Kyma dashboard locally in a docker container and open it directly in a web browser.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runDashboardStop())
		}}

	return cmd
}

func runDashboardStop() clierror.Error {
	//To be implemented in a future PR
	return nil
}
