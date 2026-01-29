package dashboard

import (
	"github.com/kyma-project/cli.v3/internal/busola"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type dashboardStopConfig struct {
	*cmdcommon.KymaConfig
	containerName string
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
			clierror.Check(runDashboardStop(&cfg))
		}}

	cmd.Flags().StringVar(&cfg.containerName, "container-name", "kyma-dashboard", `Specifies the name of the local container to stop.`)

	return cmd
}

func runDashboardStop(cfg *dashboardStopConfig) clierror.Error {
	stopper := busola.NewStopper(cfg.KymaConfig, cfg.containerName)
	if err := stopper.Stop(cfg.KymaConfig); err != nil {
		return clierror.New("failed to stop container " + cfg.containerName)
	}
	return nil
}
