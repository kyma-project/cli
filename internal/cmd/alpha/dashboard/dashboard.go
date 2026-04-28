package dashboard

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli.v3/internal/busola"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type dashboardConfig struct {
	*cmdcommon.KymaConfig
	port           string
	containerName  string
	containerId    string
	verbose        bool
	kubeconfigPath string
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

	cmd.Flags().StringVarP(&cfg.port, "port", "p", "8000", `Specifies the port on which the local dashboard will be exposed.`)
	cmd.Flags().StringVar(&cfg.containerName, "container-name", "kyma-dashboard", `Specifies the name of the local container.`)
	cmd.Flags().StringVar(&cfg.containerId, "container-id", "kyma-dashboard", `Specifies the id of the local container.`)
	cmd.Flags().BoolVarP(&cfg.verbose, "verbose", "v", true, `Enables verbose output with detailed logs.`)
	cmd.Flags().StringVar(&cfg.kubeconfigPath, "kubeconfig", "", `Path to the Kyma kubeconfig file.`)

	cmd.AddCommand(NewDashboardStartCMD(kymaConfig))
	cmd.AddCommand(NewDashboardStopCMD(kymaConfig))

	return cmd
}

func runDashboard(cfg *dashboardConfig) clierror.Error {
	dash, err := busola.New(
		cfg.containerName,
		cfg.port,
		cfg.containerId,
		cfg.verbose,
		cfg.kubeconfigPath,
	)

	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to initialize docker client"))
	}

	kubeconfig, err := getBusolaKubeconfig(cfg.kubeconfigPath)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to load kubeconfig"))
	}

	if err = dash.Start(kubeconfig); err != nil {
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

func getBusolaKubeconfig(kubeconfigPath string) (*api.Config, error) {
	if kubeconfigPath == "" {
		return nil, nil
	}

	if _, err := os.Stat(kubeconfigPath); err != nil {
		return nil, fmt.Errorf("kubeconfig file not found at %q", kubeconfigPath)
	}

	pathOptions := clientcmd.NewDefaultPathOptions()
	pathOptions.LoadingRules.ExplicitPath = kubeconfigPath

	cfg, err := pathOptions.GetStartingConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
