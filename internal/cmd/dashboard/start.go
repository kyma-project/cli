package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type dashboardStartConfig struct {
	*cmdcommon.KymaConfig
	port           string
	containerName  string
	kubeconfigPath string
	verbose        bool
}

func NewDashboardStartCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := dashboardStartConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "start [flags]",
		Short: "Run Kyma dashboard locally",
		Long:  `Use this command to run the Kyma dashboard locally in a docker container and open it directly in a web browser.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runDashboardStart(&dashboardStartConfig{}))
		}}

	cmd.Flags().StringVarP(&cfg.port, "port", "p", "3001", `Specify the port on which the local dashboard will be exposed.`)
	cmd.Flags().StringVar(&cfg.containerName, "container-name", "kyma-dashboard", `Specify the name of the local container.`)

	return cmd
}

func runDashboardStart(cfg *dashboardStartConfig) clierror.Error {
	//To be implemented in a future PR
	return nil
}
