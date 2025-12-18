package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/busola"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type dashboardStopConfig struct {
	*cmdcommon.KymaConfig
	port          string
	containerName string
	verbose       bool
	open          bool
}

func NewDashboardStopCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := dashboardStopConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "stop [flags]",
		Short: `Terminates the locally running Kyma dashboard.`,
		Long:  `Use this command to terminate the locally running Kyma dashboard in a Docker container.`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runDashboardStop(cfg))
		}}

	return cmd
}

func runDashboardStop(cfg dashboardStopConfig) clierror.Error {
	dash, err := busola.New(
		cfg.containerName,
		cfg.port,
		cfg.verbose,
	)

	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to initialize docker client"))
	}

	if err = dash.Stop(cfg.Ctx); err != nil {
		return clierror.Wrap(err, clierror.New("failed to start kyma dashboard"))
	}

	return nil
}
