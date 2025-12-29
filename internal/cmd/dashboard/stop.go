package dashboard

import (
	"github.com/docker/docker/api/types/container"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/docker"
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
	cli, err := docker.NewClient()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to initialize docker client"))
	}

	err = cli.ContainerStop(cfg.Ctx, cfg.containerName, container.StopOptions{})
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to delete container "+cfg.containerName))
	}

	return nil
}
