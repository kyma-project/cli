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

	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", o.getDefaultWorkspacePath(), "Path used to download Kyma sources.")
	cobraCmd.Flags().StringVarP(&o.ResourcePath, "resources", "r", o.getDefaultResourcePath(), "Path to a KYMA resource directory.")
	cobraCmd.Flags().StringVarP(&o.ComponentsListFile, "components", "c", o.getDefaultComponentsListFile(), "Path to a KYMA components file.")
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

	installationCfg := installConfig.Config{
		WorkersCount:                  cmd.opts.WorkersCount,
		CancelTimeout:                 cmd.opts.CancelTimeout,
		QuitTimeout:                   cmd.opts.QuitTimeout,
		HelmTimeoutSeconds:            int(cmd.opts.HelmTimeout.Seconds()),
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           cli.GetLogFunc(cmd.Verbose),
		ComponentsListFile:            cmd.opts.ComponentsListFile,
		CrdPath:                       filepath.Join(cmd.opts.ResourcePath, "cluster-essentials", "files"),
		ResourcePath:                  cmd.opts.ResourcePath,
		Version:                       "-",
	}

	var updateCh chan deployment.ProcessUpdate
	if cmd.Verbose {
		defer log.Println("Kyma uninstalled!")
	} else {
		asyncUI := asyncui.AsyncUI{StepFactory: &cmd.Factory}
		updateCh, err = asyncUI.Start()
		if err != nil {
			return err
		}
		defer asyncUI.Stop() // stop receiving update-events and wait until UI rendering is finished
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
