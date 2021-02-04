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
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/asyncui"
	"github.com/kyma-project/cli/pkg/deploy"
	"github.com/magiconair/properties"
	"github.com/spf13/cobra"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
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

	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", defaultWorkspacePath, "Path used to download Kyma sources.")
	cobraCmd.Flags().StringVarP(&o.ComponentsFile, "components", "c", defaultComponentsFile, "Path to the components file.")
	cobraCmd.Flags().StringVarP(&o.OverridesFile, "values-file", "f", "", "Path to a JSON or YAML file with configuration values.")
	cobraCmd.Flags().StringSliceVarP(&o.Overrides, "value", "", []string{}, "Set a configuration value (e.g. --value component.key='the value').")
	cobraCmd.Flags().DurationVarP(&o.CancelTimeout, "cancel-timeout", "", 900*time.Second, "Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client.")
	cobraCmd.Flags().DurationVarP(&o.QuitTimeout, "quit-timeout", "", 1200*time.Second, "Time after which the deployment is aborted. Worker goroutines may still be working in the background. This value must be greater than the value for cancel-timeout.")
	cobraCmd.Flags().DurationVarP(&o.HelmTimeout, "helm-timeout", "", 360*time.Second, "Timeout for the underlying Helm client.")
	cobraCmd.Flags().IntVar(&o.WorkersCount, "workers-count", 4, "Number of parallel workers used for the deployment.")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", localKymaDevDomain, "Domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSCrtFile, "tls-crt", "", "", "TLS certificate file for the domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSKeyFile, "tls-key", "", "", "TLS key file for the domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", defaultSource, `Installation source.
	- To use a specific release, write "kyma alpha deploy --source=1.17.1".
	- To use the master branch, write "kyma alpha deploy --source=master".
	- To use a commit, write "kyma alpha deploy --source=34edf09a".
	- To use a pull request, write "kyma alpha deploy --source=PR-9486".
	- To use the local sources, write "kyma alpha deploy --source=local".`)
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "",
		fmt.Sprintf("Kyma deployment profile. Supported profiles are: \"%s\".", strings.Join(kymaProfiles, "\", \"")))
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error

	// verify input parameters
	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	// initialize Kubernetes client
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	// initialize UI
	var ui asyncui.AsyncUI
	if !cmd.Verbose { //use async UI only if not in verbose mode
		ui = asyncui.AsyncUI{StepFactory: &cmd.Factory}
		if err := ui.Start(); err != nil {
			return err
		}
		defer ui.Stop()
	}

	// only download if not from local sources
	if cmd.opts.Source != localSource {
		if err := cmd.isCompatibleVersion(); err != nil {
			return err
		}

		//if workspace already exists ask user for deletion-approval
		_, err := os.Stat(cmd.opts.WorkspacePath)
		approvalRequired := !os.IsNotExist(err)

		if err := deploy.CloneSources(&cmd.Factory, cmd.opts.WorkspacePath, cmd.opts.Source); err != nil {
			return err
		}

		// delete workspace folder
		if approvalRequired && !cmd.avoidUserInteraction() {
			userApprovalStep := cmd.NewStep("Workspace folder exists")
			if userApprovalStep.PromptYesNo(fmt.Sprintf("Delete workspace folder '%s' after Kyma deployment?", cmd.opts.WorkspacePath)) {
				defer os.RemoveAll(cmd.opts.WorkspacePath)
			}
			userApprovalStep.Success()
		} else {
			defer os.RemoveAll(cmd.opts.WorkspacePath)
		}

	}

	err = cmd.deployKyma(ui)
	if err == nil {
		cmd.showSuccessMessage()
	}
	return err
}

func (cmd *command) isCompatibleVersion() error {
	compCheckStep := cmd.NewStep("Verifying Kyma version compatibility")
	provider := metadata.New(cmd.K8s.Static())
	clusterMetadata, err := provider.ReadKymaMetadata()
	if err != nil {
		return fmt.Errorf("Cannot get Kyma cluster version due to error: %v", err)
	}

	if clusterMetadata.Version == "" { //Kyma seems not to be installed
		compCheckStep.Successf("No previous Kyma version found")
		return nil
	}

	var compCheckFailed bool
	if clusterMetadata.Version == cmd.opts.Source {
		compCheckStep.Failuref("Current and next Kyma version are equal: %s", clusterMetadata.Version)
		compCheckFailed = true
	}
	if err := checkCompatibility(clusterMetadata.Version, cmd.opts.Source); err != nil {
		compCheckStep.Failuref("Cannot check compatibility between version '%s' and '%s'. This might cause errors - do you want to proceed anyway?", clusterMetadata.Version, cmd.opts.Source)
		compCheckFailed = true
	}
	if !compCheckFailed {
		compCheckStep.Success()
		return nil
	}

	//seemless upgrade unnecessary or cannot be warrantied - aks user for approval
	qUpgradeIncompStep := cmd.NewStep("Continue Kyma upgrade")
	if cmd.avoidUserInteraction() || qUpgradeIncompStep.PromptYesNo("Do you want to proceed with the upgrade? ") {
		qUpgradeIncompStep.Success()
		return nil
	}
	qUpgradeIncompStep.Failure()
	return fmt.Errorf("Upgrade stopped by user")
}

func (cmd *command) deployKyma(ui asyncui.AsyncUI) error {
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
		ComponentsListFile:            cmd.opts.ResolveComponentsFile(),
		CrdPath:                       filepath.Join(resourcePath, "cluster-essentials", "files"),
		ResourcePath:                  resourcePath,
		Version:                       cmd.opts.Source,
	}

	overrides, err := cmd.overrides()
	if err != nil {
		return err
	}

	// if an AsyncUI is used, get channel for update events
	var updateCh chan<- deployment.ProcessUpdate
	if ui.IsRunning() {
		updateCh, err = ui.UpdateChannel()
		if err != nil {
			return err
		}
	}

	installer, err := deployment.NewDeployment(installationCfg, overrides, cmd.K8s.Static(), updateCh)
	if err != nil {
		return err
	}

	return installer.StartKymaDeployment()
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
			return overrides, fmt.Errorf("Override has wrong format: Provide overrides in 'key=value' format")
		}

		// process key-value pair
		for _, key := range keyValuePairs.Keys() {
			value, ok := keyValuePairs.Get(key)
			if !ok || value == "" {
				return overrides, fmt.Errorf("Cannot read value of override '%s'", key)
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
	certProvided, err := cmd.opts.tlsCertAndKeyProvided()
	if err != nil {
		return err
	}
	if certProvided {
		// use encoded TLS key
		tlsKeyEnc, err := cmd.opts.tlsKeyEnc()
		if err != nil {
			return err
		}
		globalOverrides["tlsKey"] = tlsKeyEnc

		// use encoded TLS crt
		tlsCrtEnc, err := cmd.opts.tlsCrtEnc()
		if err != nil {
			return err
		}
		globalOverrides["tlsCrt"] = tlsCrtEnc
	} else {
		// use default TLS cert
		globalOverrides["tlsKey"] = defaultTLSKeyEnc
		globalOverrides["tlsCrt"] = defaultTLSCrtEnc
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
		return comp, latestOverrideMap, fmt.Errorf("Override key must contain at least the chart name "+
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
		return comp, latestOverrideMap, fmt.Errorf("Failed to extract overrides map from '%s=%s'", key, value)
	}

	return comp, latestOverrideMap, nil
}

func (cmd *command) showSuccessMessage() {
	var err error
	logFunc := cli.LogFunc(cmd.Verbose)

	fmt.Println("Kyma successfully installed.")

	tlsProvided, err := cmd.opts.tlsCertAndKeyProvided()
	if err != nil {
		logFunc("%s", err)
	}

	// show cert installation hint only for local Kyma domain and if user isn't providing a custom cert
	if (cmd.opts.Domain == localKymaDevDomain) && !tlsProvided {
		if err = cmd.storeCrtAsFile(); err != nil {
			logFunc("%s", err)
		}
		fmt.Println(`
Generated self signed TLS certificate should be trusted in your system.

  * On Mac Os X, execute this command:
   
    sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt

  * On Windows, follow the steps described here:

    https://support.globalsign.com/ssl/ssl-certificates-installation/import-and-export-certificate-microsoft-windows

This is a one time operation (you can skip this step if you did it before).`)
	}

	adminPw, err := cmd.adminPw()
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

func (cmd *command) adminPw() (string, error) {
	secret, err := cmd.K8s.Static().CoreV1().Secrets("kyma-system").Get(context.Background(), "admin-user", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data["password"]), nil
}

//avoidUserInteraction returns true if user won't provide input
func (cmd *command) avoidUserInteraction() bool {
	return cmd.NonInteractive || cmd.CI
}
