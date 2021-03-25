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
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/internal/trust"
	"github.com/kyma-project/cli/pkg/asyncui"
	"github.com/kyma-project/cli/pkg/deploy"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/magiconair/properties"
	"github.com/spf13/cobra"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/metadata"
)

type command struct {
	opts *Options
	cli.Command
	duration time.Duration
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
	cobraCmd.Flags().BoolVarP(&o.Atomic, "atomic", "a", false, "Set --atomic=true to use atomic deployment, which rolls back any component that could not be installed successfully.")
	cobraCmd.Flags().StringVarP(&o.ComponentsFile, "components", "c", defaultComponentsFile, "Path to the components file.")
	cobraCmd.Flags().StringSliceVarP(&o.OverridesFiles, "values-file", "f", []string{}, "Path to a JSON or YAML file with configuration values.")
	cobraCmd.Flags().StringSliceVarP(&o.Overrides, "value", "", []string{}, "Set a configuration value (e.g. --value component.key='the value').")
	cobraCmd.Flags().DurationVarP(&o.CancelTimeout, "cancel-timeout", "", 900*time.Second, "Time after which the workers' context is canceled. Any pending worker goroutines that are blocked by a Helm client will continue.")
	cobraCmd.Flags().DurationVarP(&o.QuitTimeout, "quit-timeout", "", 1200*time.Second, "Time after which the deployment is aborted. Worker goroutines may still be working in the background. This value must be greater than the value for cancel-timeout.")
	cobraCmd.Flags().DurationVarP(&o.HelmTimeout, "helm-timeout", "", 360*time.Second, "Timeout for the underlying Helm client.")
	cobraCmd.Flags().IntVar(&o.WorkersCount, "workers-count", 4, "Number of parallel workers used for the deployment.")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", "", "Custom domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSCrtFile, "tls-crt", "", "", "TLS certificate file for the domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSKeyFile, "tls-key", "", "", "TLS key file for the domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", defaultSource, `Installation source.
	- To use a specific release, write "kyma alpha deploy --source=1.17.1".
	- To use the master branch, write "kyma alpha deploy --source=master".
	- To use a commit, write "kyma alpha deploy --source=34edf09a".
	- To use a pull request, write "kyma alpha deploy --source=PR-9486".
	- To use the local sources, write "kyma alpha deploy --source=local".`)
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "",
		fmt.Sprintf("Kyma deployment profile. If not specified, Kyma is installed with the default chart values. The supported profiles are: \"%s\".", strings.Join(kymaProfiles, "\", \"")))
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error

	start := time.Now()
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
			userApprovalStep := cmd.NewStep("Workspace folder already exists")
			if userApprovalStep.PromptYesNo(fmt.Sprintf("Delete workspace folder '%s' after Kyma deployment?", cmd.opts.WorkspacePath)) {
				defer os.RemoveAll(cmd.opts.WorkspacePath)
			}
			userApprovalStep.Success()
		} else {
			defer os.RemoveAll(cmd.opts.WorkspacePath)
		}

	}

	overrides, err := cmd.overrides()
	if err != nil {
		return err
	}

	err = cmd.deployKyma(ui, overrides)
	if err != nil {
		return err
	}
	cmd.duration = time.Since(start)

	// import certificate
	if err := cmd.importCertificate(); err != nil {
		return err
	}

	// print summary
	o, err := overrides.Build()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve overrides to print installation summary")
	}
	return cmd.printSummary(o)
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

func (cmd *command) deployKyma(ui asyncui.AsyncUI, overrides *deployment.OverridesBuilder) error {
	localWorkspace := cmd.opts.ResolveLocalWorkspacePath()
	resourcePath := filepath.Join(localWorkspace, "resources")
	installResourcePath := filepath.Join(localWorkspace, "installation", "resources")

	cmpFile, err := cmd.opts.ResolveComponentsFile()
	if err != nil {
		return err
	}

	installationCfg := &installConfig.Config{
		WorkersCount:                  cmd.opts.WorkersCount,
		CancelTimeout:                 cmd.opts.CancelTimeout,
		QuitTimeout:                   cmd.opts.QuitTimeout,
		HelmTimeoutSeconds:            int(cmd.opts.HelmTimeout.Seconds()),
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           cli.NewHydroformLoggerAdapter(cli.NewLogger(cmd.Verbose)),
		Profile:                       cmd.opts.Profile,
		ComponentsListFile:            cmpFile,
		ResourcePath:                  resourcePath,
		InstallationResourcePath:      installResourcePath,
		Version:                       cmd.opts.Source,
		Atomic:                        cmd.opts.Atomic,
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

func (cmd *command) overrides() (*deployment.OverridesBuilder, error) {
	ob := &deployment.OverridesBuilder{}

	// add override files
	overridesFiles, err := cmd.opts.ResolveOverridesFiles()
	if err != nil {
		return ob, err
	}
	for _, overridesFile := range overridesFiles {
		if err := ob.AddFile(overridesFile); err != nil {
			return ob, err
		}
	}

	// set global overrides which the CLI allows customer to specify using CLI params (just for UX convenience)
	if err := cmd.setGlobalOverrides(ob); err != nil {
		return ob, err
	}

	// add overrides provided as CLI params
	for _, override := range cmd.opts.Overrides {
		keyValuePairs := properties.MustLoadString(override)
		if keyValuePairs.Len() < 1 {
			return ob, fmt.Errorf("Override has wrong format: Provide overrides in 'key=value' format")
		}

		// process key-value pair
		for _, key := range keyValuePairs.Keys() {
			value, ok := keyValuePairs.Get(key)
			if !ok || value == "" {
				return ob, fmt.Errorf("Cannot read value of override '%s'", key)
			}

			comp, overridesMap, err := cmd.convertToOverridesMap(key, value)
			if err != nil {
				return ob, err
			}

			if err := ob.AddOverrides(comp, overridesMap); err != nil {
				return ob, err
			}
		}
	}

	return ob, nil
}

//setGlobalOverrides is setting global overrides to improve the UX of the CLI
func (cmd *command) setGlobalOverrides(overrides *deployment.OverridesBuilder) error {
	// add domain provided as CLI params (for UX convenience)
	globalOverrides := make(map[string]interface{})
	if cmd.opts.Domain != "" {
		globalOverrides["domainName"] = cmd.opts.Domain
	}
	// add certificate provided as CLI params (for UX convenience)
	certProvided, err := cmd.opts.tlsCertAndKeyProvided()
	if err != nil {
		return err
	}
	if certProvided {
		tlsKey, err := cmd.opts.tlsKeyEnc()
		if err != nil {
			return err
		}
		tlsCrt, err := cmd.opts.tlsCrtEnc()
		if err != nil {
			return err
		}
		globalOverrides["tlsKey"] = tlsKey
		globalOverrides["tlsCrt"] = tlsCrt
	}

	// register global overrides
	if len(globalOverrides) > 0 {
		if err := overrides.AddOverrides("global", globalOverrides); err != nil {
			return err
		}
	}

	return nil
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

//avoidUserInteraction returns true if user won't provide input
func (cmd *command) avoidUserInteraction() bool {
	return cmd.NonInteractive || cmd.CI
}

func (cmd *command) printSummary(o deployment.Overrides) error {
	provider := metadata.New(cmd.K8s.Static())
	md, err := provider.ReadKymaMetadata()
	if err != nil {
		return err
	}

	domain, ok := o.Find("global.domainName")
	if !ok {
		return errors.New("Domain not found in overrides")
	}

	var consoleURL string
	vs, err := cmd.K8s.Istio().NetworkingV1alpha3().VirtualServices("kyma-system").Get(context.Background(), "console-web", metav1.GetOptions{})
	switch {
	case k8sErrors.IsNotFound(err):
		consoleURL = "not installed"
	case err != nil:
		return err
	case vs != nil && len(vs.Spec.Hosts) > 0:
		consoleURL = fmt.Sprintf("https://%s", vs.Spec.Hosts[0])
	default:
		return errors.New("console host could not be obtained")
	}

	var email, pass string
	adm, err := cmd.K8s.Static().CoreV1().Secrets("kyma-system").Get(context.Background(), "admin-user", metav1.GetOptions{})
	switch {
	case k8sErrors.IsNotFound(err):
		break
	case err != nil:
		return err
	case adm != nil:
		email = string(adm.Data["email"])
		pass = string(adm.Data["password"])
	default:
		return errors.New("admin credentials could not be obtained")
	}

	sum := nice.Summary{
		NonInteractive: cmd.NonInteractive,
		Version:        md.Version,
		URL:            domain.(string),
		Console:        consoleURL,
		Duration:       cmd.duration,
		Email:          string(email),
		Password:       string(pass),
	}

	return sum.Print()
}

func (cmd *command) importCertificate() error {
	ca := trust.NewCertifier(cmd.K8s)

	// get cert from cluster
	cert, err := ca.Certificate()
	if err != nil {
		return err
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "kyma-*.crt")
	if err != nil {
		return errors.Wrap(err, "Cannot create temporary file for Kyma certificate")
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write(cert); err != nil {
		return errors.Wrap(err, "Failed to write the kyma certificate")
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// create a simple step to print certificate import steps without a spinner (spinner overwrites sudo prompt)
	// TODO refactor how certifier logs when the old install command is gone
	f := step.Factory{
		NonInteractive: true,
	}
	s := f.NewStep("Importing Kyma certificate")

	if err := ca.StoreCertificate(tmpFile.Name(), s); err != nil {
		return err
	}
	s.Successf("Kyma root certificate imported")
	return nil
}
