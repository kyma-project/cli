package dashboard

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/dashboard"
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
		Use:   "start",
		Short: "Runs the Kyma dashboard locally and opens it directly in a web browser.",
		Long:  `Use this command to run the Kyma dashboard locally in a docker container and open it directly in a web browser. This command only works with a local installation of Kyma.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runDashboardStart(&dashboardStartConfig{}))
		}}
	cmd.Flags().StringVarP(&cfg.port, "port", "p", "3001", `Specify the port on which the local dashboard will be exposed.`)
	cmd.Flags().StringVar(&cfg.containerName, "container-name", "kyma-dashboard", `Specify the name of the local container.`)

	return cmd
}

func runDashboardStart(cfg *dashboardStartConfig) clierror.Error {

	dash := dashboard.New(cfg.containerName, cfg.port, cfg.kubeconfigPath, cfg.verbose)

	if err := dash.Start(); err != nil {
		return clierror.Wrap(err, clierror.New("failed to start kyma dashboard"))
	}

	if err := dash.Open(""); err != nil {
		return clierror.Wrap(err, clierror.New("failed to build envs from configmap"))
	}

	if err := dash.Watch(context.Background()); err != nil {
		return clierror.Wrap(err, clierror.New("failed to start kyma dashboard"))
	}
	return nil
}
