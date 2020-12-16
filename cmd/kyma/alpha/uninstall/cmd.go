package uninstall

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
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

	cobraCmd.Flags().StringVarP(&o.ComponentsYaml, "components", "c", "", "Path to a YAML file with component list to override. (required)")
	cobraCmd.Flags().DurationVarP(&o.CancelTimeout, "cancel-timeout", "", 900*time.Second, "Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client.")
	cobraCmd.Flags().DurationVarP(&o.QuitTimeout, "quit-timeout", "", 1200*time.Second, "Time after which the uninstallation is aborted. Worker goroutines may still be working in the background. This value must be greater than the value for cancel-timeout.")
	cobraCmd.Flags().DurationVarP(&o.HelmTimeout, "helm-timeout", "", 360*time.Second, "Timeout for the underlying Helm client.")
	cobraCmd.Flags().IntVar(&o.WorkersCount, "workers-count", 4, "Number of parallel workers used for the uninstallation.")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error
	if err = cmd.validateFlags(); err != nil {
		return err
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	var componentsContent string
	if cmd.opts.ComponentsYaml != "" {
		data, err := ioutil.ReadFile(cmd.opts.ComponentsYaml)
		if err != nil {
			return fmt.Errorf("Failed to read installation CR file: %v", err)
		}
		componentsContent = string(data)
	}

	prerequisitesContent := [][]string{
		{"cluster-essentials", "kyma-system"},
		{"istio", "istio-system"},
		{"xip-patch", "kyma-installer"},
	}

	installationCfg := installConfig.Config{
		WorkersCount:                  cmd.opts.WorkersCount,
		CancelTimeout:                 cmd.opts.CancelTimeout,
		QuitTimeout:                   cmd.opts.QuitTimeout,
		HelmTimeoutSeconds:            int(cmd.opts.HelmTimeout.Seconds()),
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           log.Printf,
	}

	installer, err := deployment.NewDeployment(prerequisitesContent, componentsContent, nil, "path", installationCfg)
	if err != nil {
		return err
	}

	err = installer.StartKymaUninstallation(cmd.K8s.Static())
	if err != nil {
		return err
	}

	log.Println("Kyma uninstalled!")

	return nil
}

func (cmd *command) validateFlags() error {
	if cmd.opts.ComponentsYaml == "" {
		return fmt.Errorf("Components YAML cannot be empty")
	}
	if cmd.opts.QuitTimeout < cmd.opts.CancelTimeout {
		return fmt.Errorf("Quit timeout (%v) cannot be smaller than cancel timeout (%v)", cmd.opts.QuitTimeout, cmd.opts.CancelTimeout)
	}

	return nil
}
