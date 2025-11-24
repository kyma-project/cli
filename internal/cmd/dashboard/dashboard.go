package dashboard

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/dashboard"
	"github.com/spf13/cobra"
)

type dashboardConfig struct {
	*cmdcommon.KymaConfig
	port           string
	containerName  string
	kubeconfigPath string
	verbose        bool
}

func NewDashboardCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := dashboardConfig{
		KymaConfig: kymaConfig,
	}

	fmt.Println("testestst")
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Runs the Kyma dashboard locally and opens it directly in a web browser.",
		Long:  `Use this command to run the Kyma dashboard locally in a docker container and open it directly in a web browser. This command only works with a local installation of Kyma.`,
		Run: func(cmd *cobra.Command, args []string) {
			runDashboard(&cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.port, "port", "p", "3001", `Specify the port on which the local dashboard will be exposed.`)
	cmd.Flags().StringVar(&cfg.containerName, "container-name", "kyma-dashboard", `Specify the name of the local container.`)

	return cmd
}

func runDashboard(cfg *dashboardConfig) {
	_ = dashboard.New(cfg.containerName, cfg.port, cfg.kubeconfigPath, cfg.verbose)

	return
}
