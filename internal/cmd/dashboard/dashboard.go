package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/busola"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type dashboardConfig struct {
	*cmdcommon.KymaConfig
	port          string
	containerName string
	verbose       bool
}

func NewDashboardCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := dashboardConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "dashboard <command> [flags]",
		Short: "Manages Kyma dashboard locally.",
		Long:  `Use this command to manage Kyma dashboard locally in a Docker container.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runDashboard(&cfg))
		}}

	cmd.Flags().StringVarP(&cfg.port, "port", "p", "3001", `Specifies the port on which the local dashboard will be exposed.`)
	cmd.Flags().StringVar(&cfg.containerName, "container-name", "kyma-dashboard", `Specifies the name of the local container.`)
	cmd.Flags().BoolVarP(&cfg.verbose, "verbose", "v", true, `Enables verbose output with detailed logs.`)

	cmd.AddCommand(NewDashboardStartCMD(kymaConfig))
	cmd.AddCommand(NewDashboardStopCMD(kymaConfig))

	return cmd
}

func runDashboard(cfg *dashboardConfig) clierror.Error {
	dash, err := busola.New(
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

	if err = dash.Open(); err != nil {
		return clierror.Wrap(err, clierror.New("failed to open kyma dashboard"))
	}

	if err = dash.Watch(); err != nil {
		return clierror.Wrap(err, clierror.New("failed to watch kyma dashboard"))
	}

	return nil
}
