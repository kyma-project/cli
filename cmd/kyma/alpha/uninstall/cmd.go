package uninstall

import (
	"log"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/asyncui"
	"github.com/spf13/cobra"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstalls Kyma from a running Kubernetes cluster.",
		Long:    `Use this command to uninstall Kyma from a running Kubernetes cluster.`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"i"},
	}

	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", o.defaultWorkspacePath(), "Path used to download Kyma sources.")
	cobraCmd.Flags().StringVarP(&o.OverridesFile, "overrides", "o", "", "Path to a JSON or YAML file with parameters to override.")
	cobraCmd.Flags().DurationVarP(&o.CancelTimeout, "cancel-timeout", "", 900*time.Second, "Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client.")
	cobraCmd.Flags().DurationVarP(&o.QuitTimeout, "quit-timeout", "", 1200*time.Second, "Time after which the uninstallation is aborted. Worker goroutines may still be working in the background. This value must be greater than the value for cancel-timeout.")
	cobraCmd.Flags().DurationVarP(&o.HelmTimeout, "helm-timeout", "", 360*time.Second, "Timeout for the underlying Helm client.")
	cobraCmd.Flags().IntVar(&o.WorkersCount, "workers-count", 4, "Number of parallel workers used for the uninstallation.")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error
	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	var resourcePath = filepath.Join(cmd.opts.WorkspacePath, "kyma", "resources")
	installationCfg := installConfig.Config{
		WorkersCount:                  cmd.opts.WorkersCount,
		CancelTimeout:                 cmd.opts.CancelTimeout,
		QuitTimeout:                   cmd.opts.QuitTimeout,
		HelmTimeoutSeconds:            int(cmd.opts.HelmTimeout.Seconds()),
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           cli.LogFunc(cmd.Verbose),
		ComponentsListFile:            filepath.Join(cmd.opts.WorkspacePath, "kyma", "installation", "resources", "components.yaml"),
		CrdPath:                       filepath.Join(resourcePath, "cluster-essentials", "files"),
		ResourcePath:                  resourcePath,
		Version:                       "-",
	}

	var ui asyncui.AsyncUI
	if cmd.Verbose {
		defer log.Println("Kyma uninstalled!")
	} else {
		asyncUI := asyncui.AsyncUI{StepFactory: &cmd.Factory}
		if err = asyncUI.Start(); err != nil {
			return err
		}
		defer asyncUI.Stop() // stop receiving update-events and wait until UI rendering is finished
	}

	// if an AsyncUI is used, get channel for update events
	var updateCh chan<- deployment.ProcessUpdate
	if ui.IsRunning() {
		updateCh, err = ui.UpdateChannel()
		if err != nil {
			return err
		}
	}

	installer, err := deployment.NewDeployment(installationCfg, deployment.Overrides{}, cmd.K8s.Static(), updateCh)
	if err != nil {
		return err
	}

	err = installer.StartKymaUninstallation()
	if err != nil {
		return err
	}

	return nil
}
