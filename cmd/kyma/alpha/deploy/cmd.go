package deploy

import (
	"fmt"
	"log"
	"path/filepath"
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

	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", o.defaultWorkspacePath(), "Path used to download Kyma sources.")
	cobraCmd.Flags().StringVarP(&o.ComponentsFile, "components", "c", o.defaultComponentsFile(), "Path to the components file.")
	cobraCmd.Flags().StringVarP(&o.OverridesFile, "overrides", "o", "", "Path to a JSON or YAML file with parameters to override.")
	cobraCmd.Flags().DurationVarP(&o.CancelTimeout, "cancel-timeout", "", 900*time.Second, "Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client.")
	cobraCmd.Flags().DurationVarP(&o.QuitTimeout, "quit-timeout", "", 1200*time.Second, "Time after which the deployment is aborted. Worker goroutines may still be working in the background. This value must be greater than the value for cancel-timeout.")
	cobraCmd.Flags().DurationVarP(&o.HelmTimeout, "helm-timeout", "", 360*time.Second, "Timeout for the underlying Helm client.")
	cobraCmd.Flags().IntVar(&o.WorkersCount, "workers-count", 4, "Number of parallel workers used for the deployment.")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", o.defaultDomain(), "Domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSCert, "tls-cert", "", "", "TLS certificate for the domain used for installation. The certificate must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.TLSKey, "tls-key", "", "", "TLS key for the domain used for installation. The key must be a base64-encoded value.")
	cobraCmd.Flags().StringVarP(&o.Version, "source", "s", o.defaultVersion(), `Installation source. 
	- To use the latest release, write "kyma alpha deploy --source=latest".
	- To use a specific release, write "kyma alpha deploy --source=1.17.1".
	- To use the master branch, write "kyma alpha deploy --source=master".
	- To use a commit, write "kyma alpha deploy --source=34edf09a".
	- To use a pull request, write "kyma alpha deploy --source=PR-9486".
	- To use the local sources, write "kyma alpha deploy --source=local".`)
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "",
		fmt.Sprintf("Kyma deployment profile. Supported profiles are: \"%s\".", strings.Join(o.profiles(), "\", \"")))
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

	//to add another stop to the UI, just call:
	if result, _ := cmd.k3sCluster(); result {
		step := cmd.startDeploymentStep(updateCh, "This is a k3s specific step")
		cmd.stopDeploymentStep(updateCh, step, true)
	}

	return cmd.deployKyma(updateCh)
}

func (cmd *command) deployKyma(updateCh chan<- deployment.ProcessUpdate) error {
	var resourcePath = filepath.Join(cmd.opts.WorkspacePath, "kyma", "resources")

	installationCfg := installConfig.Config{
		WorkersCount:                  cmd.opts.WorkersCount,
		CancelTimeout:                 cmd.opts.CancelTimeout,
		QuitTimeout:                   cmd.opts.QuitTimeout,
		HelmTimeoutSeconds:            int(cmd.opts.HelmTimeout.Seconds()),
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           cli.LogFunc(cmd.Verbose),
		Profile:                       cmd.opts.Profile,
		ComponentsListFile:            cmd.opts.ComponentsFile,
		CrdPath:                       filepath.Join(resourcePath, "cluster-essentials", "files"),
		ResourcePath:                  resourcePath,
		Version:                       cmd.opts.Version,
		//TLSCert:                       cmd.opts.TLSCert,
		//TLSKey:                        cmd.opts.TLSKey,
		//Domain:                        cmd.ops.Domain,
	}

	overrides, err := cmd.overrides()
	if err != nil {
		return err
	}

	installer, err := deployment.NewDeployment(installationCfg, overrides, cmd.K8s.Static(), updateCh)
	if err != nil {
		return err
	}

	return installer.StartKymaDeployment()
}

func (cmd *command) overrides() (deployment.Overrides, error) {
	overrides := deployment.Overrides{}
	if cmd.opts.OverridesFile != "" {
		if err := overrides.AddFile(cmd.opts.OverridesFile); err != nil {
			return overrides, err
		}
	}
	return overrides, nil
}

func (cmd *command) startDeploymentStep(updateCh chan<- deployment.ProcessUpdate, step string) deployment.InstallationPhase {
	comp := components.KymaComponent{}
	phase := deployment.InstallationPhase(step)
	if updateCh == nil {
		cli.LogFunc(cmd.opts.Verbose)(step)
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
		cli.LogFunc(cmd.opts.Verbose)("%s finished successfully: %t", string(phase), success)
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

// k3sCluster is a helper function to figure out whether Kyma should be installed on K3s
func (cmd *command) k3sCluster() (bool, error) {
	clInfo := clusterinfo.New(cmd.K8s.Static())
	if err := clInfo.Read(); err != nil {
		return false, err
	}
	provider, err := clInfo.Provider()
	if err != nil {
		return false, err
	}
	return (provider != clusterinfo.ClusterProviderK3s), nil
}
