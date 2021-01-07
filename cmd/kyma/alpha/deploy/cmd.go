package deploy

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
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
		Use:     "deploy",
		Short:   "Deploys Kyma on a running Kubernetes cluster.",
		Long:    `Use this command to deploy Kyma on a running Kubernetes cluster.`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"d"},
	}

	cobraCmd.Flags().StringVarP(&o.OverridesYaml, "overrides", "o", "", "Path to a YAML file with parameters to override.")
	cobraCmd.Flags().StringVarP(&o.ComponentsYaml, "components", "c", "", "Path to a YAML file with component list to override. (required)")
	cobraCmd.Flags().StringVarP(&o.ResourcesPath, "resources", "r", "", "Path to Kyma resources folder. (required)")
	cobraCmd.Flags().DurationVarP(&o.CancelTimeout, "cancel-timeout", "", 900*time.Second, "Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client.")
	cobraCmd.Flags().DurationVarP(&o.QuitTimeout, "quit-timeout", "", 1200*time.Second, "Time after which the deployment is aborted. Worker goroutines may still be working in the background. This value must be greater than the value for cancel-timeout.")
	cobraCmd.Flags().DurationVarP(&o.HelmTimeout, "helm-timeout", "", 360*time.Second, "Timeout for the underlying Helm client.")
	cobraCmd.Flags().IntVar(&o.WorkersCount, "workers-count", 4, "Number of parallel workers used for the deployment.")
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "evaluation",
		fmt.Sprintf("Kyma deployment profile Supported profiles are: %s", strings.Join(o.GetProfiles(), ", ")))
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error
	if err = cmd.opts.ValidateFlags(); err != nil {
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

	var overridesContent []string
	if cmd.opts.OverridesYaml != "" {
		data, err := ioutil.ReadFile(cmd.opts.OverridesYaml)
		if err != nil {
			return fmt.Errorf("Failed to read installation CR file: %v", err)
		}
		overridesContent = append(overridesContent, string(data))
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
		Log:                           cli.GetLogFunc(cmd.Verbose),
		Profile:                       cmd.opts.Profile,
	}

	var updateCh chan deployment.ProcessUpdate
	if cmd.Verbose {
		defer log.Println("Kyma deployed!")
	} else {
		asyncUI := asyncui.AsyncUI{StepFactory: &cmd.Factory}
		updateCh, err = asyncUI.Start()
		if err != nil {
			return err
		}
		defer asyncUI.Stop() // stop receiving update-events and wait until UI rendering is finished
	}

	installer, err := deployment.NewDeployment(prerequisitesContent, componentsContent, overridesContent, cmd.opts.ResourcesPath, installationCfg, updateCh)
	if err != nil {
		return err
	}

	err = installer.StartKymaDeployment(cmd.K8s.Static())
	if err != nil {
		return err
	}

	return nil
}
