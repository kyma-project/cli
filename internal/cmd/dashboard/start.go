package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/docker"
	"github.com/spf13/cobra"
)

type dashboardStartConfig struct {
	*cmdcommon.KymaConfig
	port          string
	containerName string
	verbose       bool
	open          bool
}

func NewDashboardStartCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := dashboardStartConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "start [flags]",
		Short: `Runs Kyma dashboard locally.`,
		Long:  `Use this command to run Kyma dashboard locally in a Docker container and open it directly in a web browser.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runDashboardStart(&cfg))
		}}

	cmd.Flags().StringVarP(&cfg.port, "port", "p", "3001", `Specify the port on which the local dashboard will be exposed.`)
	cmd.Flags().StringVar(&cfg.containerName, "container-name", "kyma-dashboard", `Specify the name of the local container.`)
	cmd.Flags().BoolVarP(&cfg.verbose, "verbose", "v", true, `Enable verbose output with detailed logs.`)
	cmd.Flags().BoolVarP(&cfg.open, "open", "o", false, `Specify if the browser should open after executing the command.`)

	return cmd
}

func runDashboardStart(cfg *dashboardStartConfig) clierror.Error {
	dash, err := docker.New(
		cfg.containerName,
		cfg.port,
		cfg.verbose,
	)

	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to initialize docker client"))
	}

	if err = dash.Start(); err != nil {
		return clierror.Wrap(err, clierror.New("failed to start kyma dashboard"))
	}

	if cfg.open {
		err = dash.Open("")
	}
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to open kyma dashboard"))
	}

	return nil
}
