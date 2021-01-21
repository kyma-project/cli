package deploy

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/asyncui"
	"github.com/kyma-project/cli/pkg/git"
	"github.com/magiconair/properties"
	"github.com/spf13/cobra"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
)

const (
	kymaURL      = "https://github.com/kyma-project/kyma"
	localVersion = "local"
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
	cobraCmd.Flags().StringVarP(&o.OverridesFile, "values-file", "f", "", "Path to a JSON or YAML file with configuration values.")
	cobraCmd.Flags().StringSliceVarP(&o.Overrides, "value", "", []string{}, "Set a configuration value (e.g. --value component.key='the value').")
	cobraCmd.Flags().DurationVarP(&o.CancelTimeout, "cancel-timeout", "", 900*time.Second, "Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client.")
	cobraCmd.Flags().DurationVarP(&o.QuitTimeout, "quit-timeout", "", 1200*time.Second, "Time after which the deployment is aborted. Worker goroutines may still be working in the background. This value must be greater than the value for cancel-timeout.")
	cobraCmd.Flags().DurationVarP(&o.HelmTimeout, "helm-timeout", "", 360*time.Second, "Timeout for the underlying Helm client.")
	cobraCmd.Flags().IntVar(&o.WorkersCount, "workers-count", 4, "Number of parallel workers used for the deployment.")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", LocalKymaDevDomain, "Domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSCrt, "tls-crt", "", "", "TLS certificate for the domain used for installation. The certificate must be a base64-encoded value.")
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
		//TODO: check for errors instead showing success message blindly - depends on https://github.com/kyma-incubator/hydroform/issues/170
		defer cmd.showSuccessMessage()
	} else {
		asyncUI := asyncui.AsyncUI{StepFactory: &cmd.Factory}
		updateCh, err = asyncUI.Start()
		if err != nil {
			return err
		}
		defer func() {
			asyncUI.Stop() // stop receiving update-events and wait until UI rendering is finished
			if !asyncUI.Failed {
				cmd.showSuccessMessage()
			}
		}()
	}

	// only download if not from local sources
	if cmd.opts.Version != localVersion {
		step := cmd.startDeploymentStep(updateCh, "Downloading Kyma")

		rev, err := git.ResolveRevision(kymaURL, cmd.opts.Version)
		if err != nil {
			cmd.stopDeploymentStep(updateCh, step, false)
			return err
		}

		if err := git.CloneRevision(kymaURL, cmd.opts.WorkspacePath, rev); err != nil {
			cmd.stopDeploymentStep(updateCh, step, false)
			return err
		}
		// only delete sources if clone was successful
		defer os.RemoveAll(cmd.opts.WorkspacePath)
		cmd.stopDeploymentStep(updateCh, step, true)
	}

	return cmd.deployKyma(updateCh)
}

func (cmd *command) showSuccessMessage() {
	var err error
	logFunc := cli.LogFunc(cmd.Verbose)

	isLocal, err := cmd.isLocalSetup()
	if err != nil {
		logFunc("%s", err)
	}

	fmt.Println("Kyma successfully installed.")

	if isLocal {
		if err = cmd.storeCrtAsFile(); err != nil {
			logFunc("%s", err)
		}
		fmt.Println(`
Generated self signed TLS certificate should be trusted in your system.

  * On Mac Os X execute this command:
   
    sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt

  * On Windows please follow the steps described here:

    https://support.globalsign.com/ssl/ssl-certificates-installation/import-and-export-certificate-microsoft-windows

This is a one time operation (you can skip this step if you did it before).`)
	}

	adminPw, err := cmd.getAdminPw()
	if err != nil {
		logFunc("%s", err)
	}
	fmt.Printf(`
Kyma Console Url: %s
User: admin@kyma.cx
Password: %s

`, fmt.Sprintf("https://console.%s", cmd.opts.Domain), adminPw)
}

func (cmd *command) storeCrtAsFile() error {
	secret, err := cmd.K8s.Static().CoreV1().Secrets("istio-system").Get(context.Background(), "kyma-gateway-certs", metav1.GetOptions{})
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile("kyma.crt", secret.Data["cert"], 0600); err != nil {
		return err
	}
	return nil
}

func (cmd *command) getAdminPw() (string, error) {
	secret, err := cmd.K8s.Static().CoreV1().Secrets("kyma-system").Get(context.Background(), "admin-user", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data["password"]), nil
}

func (cmd *command) deployKyma(updateCh chan<- deployment.ProcessUpdate) error {
	var resourcePath = filepath.Join(cmd.opts.WorkspacePath, "resources")

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
	}

	overrides, err := cmd.overrides()
	if err != nil {
		return err
	}

	if err := cmd.configureCoreDNS(updateCh); err != nil {
		return err
	}

	installer, err := deployment.NewDeployment(installationCfg, overrides, cmd.K8s.Static(), updateCh)
	if err != nil {
		return err
	}

	return installer.StartKymaDeployment()
}

func (cmd *command) configureCoreDNS(updateCh chan<- deployment.ProcessUpdate) error {
	if cmd.opts.Domain == LocalKymaDevDomain { //patch Kubernetes DNS system when using "local.kyma.dev" as domain name
		step := cmd.startDeploymentStep(updateCh, "Configure Kubernetes DNS to support Kyma local dev domain")
		err := ConfigureCoreDNS(cmd.K8s.Static())
		cmd.stopDeploymentStep(updateCh, step, (err == nil))
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd *command) overrides() (deployment.Overrides, error) {
	overrides := deployment.Overrides{}

	// add override files
	if cmd.opts.OverridesFile != "" {
		if err := overrides.AddFile(cmd.opts.OverridesFile); err != nil {
			return overrides, err
		}
	}

	if err := cmd.setGlobalOverrides(&overrides); err != nil {
		return overrides, err
	}

	// add overrides provided as CLI params
	for _, override := range cmd.opts.Overrides {
		keyValuePairs := properties.MustLoadString(override)
		if keyValuePairs.Len() < 1 {
			return overrides, fmt.Errorf("Override has wrong format: please provide overrides in 'key=value' format")
		}

		// process key-value pair
		for _, key := range keyValuePairs.Keys() {
			value, ok := keyValuePairs.Get(key)
			if !ok || value == "" {
				return overrides, fmt.Errorf("Could not read value of override '%s'", key)
			}

			comp, overridesMap, err := cmd.convertToOverridesMap(key, value)
			if err != nil {
				return overrides, err
			}

			if err := overrides.AddOverrides(comp, overridesMap); err != nil {
				return overrides, err
			}
		}
	}

	return overrides, nil
}

func (cmd *command) setGlobalOverrides(overrides *deployment.Overrides) error {
	globalOverrides := make(map[string]interface{})
	globalOverrides["isLocalEnv"] = false //DEPRECATED - 'isLocalEnv' will be removed soon
	globalOverrides["domainName"] = cmd.opts.Domain
	if cmd.opts.tlsCertAndKeyProvided() {
		globalOverrides["tlsKey"] = cmd.opts.TLSKey
		globalOverrides["tlsCrt"] = cmd.opts.TLSCrt
	} else {
		globalOverrides["tlsKey"] = cmd.opts.defaultTLSKey()
		globalOverrides["tlsCrt"] = cmd.opts.defaultTLSCrt()
	}

	// ingress settings
	ingressOverrides := make(map[string]interface{})
	ingressOverrides["domainName"] = cmd.opts.Domain
	globalOverrides["ingress"] = ingressOverrides

	// environment overrides
	envOverrides := make(map[string]interface{})
	envOverrides["gardener"] = false
	globalOverrides["environment"] = envOverrides //DEPRECATED - 'environment' will be removed soon

	return overrides.AddOverrides("global", globalOverrides)
}

// convertToOverridesMap parses the override key and converts it into an nested map.
// First element of the key is returned as component name, all other elements are used as key/sub-key in the nested map.
func (cmd *command) convertToOverridesMap(key, value string) (string, map[string]interface{}, error) {
	var comp string
	var latestOverrideMap map[string]interface{}

	keyTokens := strings.Split(key, ".")
	if len(keyTokens) < 2 {
		return comp, latestOverrideMap, fmt.Errorf("Override key has to contain at least the chart name "+
			"and one override: chart.override[.suboverride]=value (given was '%s=%s')", key, value)
	}

	// first token in key is the chart name
	comp = keyTokens[0]

	// use the remaining key-tokens to build the nested overrides map
	// processing starts from last element to the beginning
	for idx := range keyTokens[1:] {
		overrideMap := make(map[string]interface{})     // current override-map
		overrideName := keyTokens[len(keyTokens)-1-idx] // get last token element
		if idx == 0 {
			// this is the last key-token, use it value
			overrideMap[overrideName] = value
		} else {
			// the latest override map has to become a sub-map of the current override-map
			overrideMap[overrideName] = latestOverrideMap
		}
		//set the current override map as latest override map
		latestOverrideMap = overrideMap
	}

	if len(latestOverrideMap) < 1 {
		return comp, latestOverrideMap, fmt.Errorf("Failed to extracted overrides map from '%s=%s'", key, value)
	}

	return comp, latestOverrideMap, nil
}

// isLocalSetup is a helper function to figure out whether Kyma runs locally
func (cmd *command) isLocalSetup() (bool, error) {
	clInfo := clusterinfo.New(cmd.K8s.Static())
	if err := clInfo.Read(); err != nil {
		return false, err
	}
	provider, err := clInfo.Provider()
	if err != nil {
		return false, err
	}
	return (provider == clusterinfo.ClusterProviderK3s), nil
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
