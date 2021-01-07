package deploy

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/asyncui"

	"github.com/spf13/cobra"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
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
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", o.getDefaultDomain(), "Domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSCert, "tls-cert", "", "", "TLS certificate for the domain used for installation. The certificate must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.TLSKey, "tls-key", "", "", "TLS key for the domain used for installation. The key must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", o.getDefaultVersion(), `Installation source. 
	- To use a specific release, write "kyma install --source=1.17.1".
	- To use the master branch, write "kyma install --source=master".
	- To use a commit, write "kyma install --source=34edf09a".
	- To use a pull request, write "kyma install --source=PR-9486".
	- To use the local sources, write "kyma install --source=local".`)
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "evaluation",
		fmt.Sprintf("Kyma deployment profile. Supported profiles are: %s", strings.Join(o.getProfiles(), ", ")))
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error

	// verify input parameters
	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}

	// initialize Kubernetes client
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	// initialize UI
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

	// apply changes required for local setup
	if err := cmd.finalizeK3sSetup(updateCh); err != nil {
		return err
	}

	// execute deployment
	installer, err := cmd.getInstaller(updateCh)
	if err != nil {
		return err
	}
	if err = installer.StartKymaDeployment(cmd.K8s.Static()); err != nil {
		return err
	}

	return nil
}

func (cmd *command) finalizeK3sSetup(updateCh chan<- deployment.ProcessUpdate) error {
	phase := cmd.startDeploymentStep(updateCh, "Verify cluster metadata")
	isK3sCluster, err := cmd.isK3sCluster()
	cmd.stopDeploymentStep(updateCh, phase, (err == nil))
	if err != nil {
		return err
	}
	if isK3sCluster {
		//TBD: add k3s specific steps
	}
	return nil
}

func (cmd *command) isK3sCluster() (bool, error) {
	clInfo := clusterinfo.NewClusterInfo(cmd.K8s.Static())
	if err := clInfo.Read(); err != nil {
		return false, err
	}
	provider, err := clInfo.GetProvider()
	if err != nil {
		return false, err
	}
	return (provider != "k3s"), nil
}

func (cmd *command) getInstaller(updateCh chan deployment.ProcessUpdate) (*deployment.Deployment, error) {
	componentsContent, err := cmd.getComponents()
	if err != nil {
		return nil, err
	}

	overridesContent, err := cmd.getOverrides()
	if err != nil {
		return nil, err
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
		//Source:                        cmd.opts.Source,
	}

	return deployment.NewDeployment(prerequisitesContent, componentsContent, overridesContent, cmd.opts.ResourcesPath, installationCfg, updateCh)
}

func (cmd *command) getComponents() (string, error) {
	var componentsContent string
	if cmd.opts.ComponentsYaml != "" {
		data, err := ioutil.ReadFile(cmd.opts.ComponentsYaml)
		if err != nil {
			return "", fmt.Errorf("Failed to read installation CR file: %v", err)
		}
		componentsContent = string(data)
	}
	return componentsContent, nil
}

func (cmd *command) getOverrides() ([]string, error) {
	var overridesContent []string
	if cmd.opts.OverridesYaml != "" {
		data, err := ioutil.ReadFile(cmd.opts.OverridesYaml)
		if err != nil {
			return nil, fmt.Errorf("Failed to read installation CR file: %v", err)
		}
		overridesContent = append(overridesContent, string(data))
	}
	return overridesContent, nil
}

func (cmd *command) startDeploymentStep(updateCh chan<- deployment.ProcessUpdate, step string) deployment.InstallationPhase {
	comp := components.KymaComponent{}
	phase := deployment.InstallationPhase(step)
	if updateCh == nil {
		cli.GetLogFunc(cmd.opts.Verbose)(step)
	} else {
		updateCh <- deployment.ProcessUpdate{
			Component: comp,
			Event:     deployment.ProcessStart,
			Phase:     phase,
		}
	}
	return phase
}

func (cmd *command) stopDeploymentStep(updateCh chan<- deployment.ProcessUpdate, phase deployment.InstallationPhase, success bool) {
	comp := components.KymaComponent{}
	if updateCh == nil {
		cli.GetLogFunc(cmd.opts.Verbose)("%s finished successfully: %t", string(phase), success)
		return
	}
	if success {
		updateCh <- deployment.ProcessUpdate{
			Component: comp,
			Event:     deployment.ProcessFinished,
			Phase:     phase,
		}
	} else {
		updateCh <- deployment.ProcessUpdate{
			Component: comp,
			Event:     deployment.ProcessExecutionFailure,
			Phase:     phase,
		}
	}
}
