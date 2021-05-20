package deploy

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/internal/trust"
	"github.com/kyma-project/cli/pkg/asyncui"
	"github.com/kyma-project/cli/pkg/installation"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/magiconair/properties"
	"github.com/spf13/cobra"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/git"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/helm"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller"
)

type command struct {
	opts *Options
	cli.Command
	duration time.Duration
}

const kymaURL = "https://github.com/kyma-project/kyma"
const kyma2OverridesPath = "/resources/installation/kyma2-values.yaml"

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys Kyma on a running Kubernetes cluster.",
		Long: `Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.

Usage Examples:
  Deploy Kyma using your own domain name
    You must provide the certificate and key as files.
    If you don't have a certificate yet, you can create a self-signed certificate and key:
		openssl req -x509 -newkey rsa:4096 -keyout key.pem -out crt.pem -days 365
    Then, pass the certificate files to the deploy command:
		kyma alpha deploy --domain {DOMAIN} --tls-cert crt.pem --tls-key key.pem

  Deploy Kyma from specific source:
    - Deploy from a specific version, such as 1.19.1:
		kyma alpha deploy --source=1.19.1
    - Build Kyma from local sources and deploy on remote cluster:
		kyma alpha deploy --source=local

  Deploy Kyma with only specific components:
    You need to pass a path to a YAML file containing desired components. An example YAML file would contain:
		prerequisites:
		- name: "cluster-essentials"
		- name: "istio"
		  namespace: "istio-system"
		components:
		- name: "testing"
		- name: "xip-patch"
		- name: "istio-kyma-patch"
		- name: "dex"
    Then run:
		kyma alpha deploy --components {COMPONENTS_FILE_PATH}

  Change Kyma settings:
    To change your Kyma configuration, use the alpha deploy command and deploy the same Kyma version that you're currently using,
    just with different settings.
    - Using a settings-file:
		kyma alpha deploy --values-file {VALUES_FILE_PATH}
    - Using specific values instead of file:
		kyma deploy --value ory.hydra.deployment.resources.limits.cpu=153m \
		--value ory.hydra.deployment.resources.requests.cpu=53m

Debugging:
  The alpha commands support troubleshooting in several ways, for example:
  - To get a detailed view of the installation process, use the --verbose flag.
  - To tweak the values on a component level, use alpha deploy --components:
    Pass a components list that includes only the components you want to test
    and try out the settings that work for your installation.
  - To understand which component failed during deployment, deactivate the default atomic deployment:
		--atomic=false
    With atomic deployment active, any component that hasn't been installed successfully is rolled back,
    which may make it hard to find out what went wrong. By disabling the flag, the failed components are not rolled back.
	`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"d"},
	}

	// default value for workspace flag is set in validateFlags()
	// to avoid having the actual home directory shown in the auto-generated docs
	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", "", `Path to download Kyma sources (default "$HOME/.kyma/sources" or ".kyma-sources")`)
	cobraCmd.Flags().BoolVarP(&o.Atomic, "atomic", "a", false, "Set --atomic=true to use atomic deployment, which rolls back any component that could not be installed successfully.")
	// default value for components-file flag is set in validateFlags()
	// to avoid having the actual home directory shown in the auto-generated docs
	cobraCmd.Flags().StringVarP(&o.ComponentsFile, "components-file", "c", "", `Path to the components file (default "$HOME/.kyma/sources/installation/resources/components.yaml" or ".kyma-sources/installation/resources/components.yaml")`)
	cobraCmd.Flags().StringSliceVarP(&o.Components, "component", "", []string{}, "Provide one or more components to deploy (e.g. --component componentName@namespace)")
	cobraCmd.Flags().StringSliceVarP(&o.OverridesFiles, "values-file", "f", []string{}, "Path(s) to one or more JSON or YAML files with configuration values")
	cobraCmd.Flags().StringSliceVarP(&o.Overrides, "value", "", []string{}, "Set configuration values. Can specify one or more values, also as a comma-separated list (e.g. --value component.a='1' --value component.b='2' or --value component.a='1',component.b='2').")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 20*time.Minute, "Maximum time for the deployment")
	cobraCmd.Flags().DurationVarP(&o.TimeoutComponent, "timeout-component", "", 6*time.Minute, "Maximum time to deploy the component")
	cobraCmd.Flags().IntVar(&o.Concurrency, "concurrency", 4, "Number of parallel processes")
	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", "", "Custom domain used for installation")
	cobraCmd.Flags().StringVarP(&o.TLSCrtFile, "tls-crt", "", "", "TLS certificate file for the domain used for installation")
	cobraCmd.Flags().StringVarP(&o.TLSKeyFile, "tls-key", "", "", "TLS key file for the domain used for installation")
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", defaultSource, `Installation source:
	- Deploy a specific release, for example: "kyma alpha deploy --source=1.17.1"
	- Deploy a specific branch of the Kyma repository on kyma-project.org: "kyma alpha deploy --source=<my-branch-name>"
	- Deploy a commit, for example: "kyma alpha deploy --source=34edf09a"
	- Deploy a pull request, for example "kyma alpha deploy --source=PR-9486"
	- Deploy the local sources: "kyma alpha deploy --source=local"`)
	setSource(cobraCmd.Flags().Changed("source"), &o.Source)
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "",
		fmt.Sprintf("Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: \"%s\".", strings.Join(kymaProfiles, "\", \"")))
	return cobraCmd
}

func setSource(isUserDefined bool, source *string) {
	if !isUserDefined {
		*source = installation.SetKymaSemVersion(*source)
	}
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

	// only download if not from local sources
	if cmd.opts.Source != localSource {
		if err := cmd.isCompatibleVersion(); err != nil {
			return err
		}

		//if workspace already exists ask user for deletion-approval
		_, err := os.Stat(cmd.opts.WorkspacePath)
		approvalRequired := !os.IsNotExist(err)

		downloadStep := cmd.NewStep(fmt.Sprintf("Downloading Kyma (%s) into workspace folder ", cmd.opts.Source))
		if err := git.CloneRepo(kymaURL, cmd.opts.WorkspacePath, cmd.opts.Source); err != nil {
			downloadStep.Failure()
			return err
		}
		downloadStep.Successf("Kyma downloaded into workspace folder")

		// delete workspace folder
		if approvalRequired && !cmd.avoidUserInteraction() {
			userApprovalStep := cmd.NewStep("Workspace folder already exists")
			if userApprovalStep.PromptYesNo(fmt.Sprintf("Delete workspace folder '%s' after Kyma deployment? ", cmd.opts.WorkspacePath)) {
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

	overrides.AddFile(filepath.Join(cmd.opts.WorkspacePath, kyma2OverridesPath))

	err = cmd.deployKyma(overrides)
	if err != nil {
		return err
	}
	cmd.duration = time.Since(start)

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
	provider, err := helm.NewKymaMetadataProvider(installConfig.KubeconfigSource{
		Path: kube.KubeconfigPath(cmd.KubeconfigPath),
	})
	if err != nil {
		return err
	}

	versionSet, err := provider.Versions()
	if err != nil {
		return errors.Wrap(err, "Cannot get installed Kyma versions due to error")
	}

	if versionSet.Empty() { //Kyma seems not to be installed
		compCheckStep.Successf("No previous Kyma version found")
		return nil
	}

	var compCheckFailed bool
	if versionSet.Count() > 1 {
		compCheckStep.Failuref("Components from multiple Kyma versions are installed (found Kyma versions '%s'). "+
			"Cannot check compatibility if components with different Kyma versions are installed.",
			strings.Join(versionSet.Names(), "', '"))
		compCheckFailed = true
	} else {
		kymaVersion := versionSet.Versions[0].Version
		if kymaVersion == cmd.opts.Source {
			compCheckStep.Failuref("Current and next Kyma version are equal: %s", kymaVersion)
			compCheckFailed = true
		}
		if err := checkCompatibility(kymaVersion, cmd.opts.Source); err != nil {
			compCheckStep.Failuref("Cannot check compatibility between version '%s' and '%s'. This might cause errors!",
				kymaVersion, cmd.opts.Source)
			compCheckFailed = true
		}
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

func (cmd *command) deployKyma(overrides *deployment.OverridesBuilder) error {
	localWorkspace := cmd.opts.ResolveLocalWorkspacePath()
	resourcePath := filepath.Join(localWorkspace, "resources")
	installResourcePath := filepath.Join(localWorkspace, "installation", "resources")

	compList, err := cmd.createCompList()
	if err != nil {
		return err
	}

	installationCfg := &installConfig.Config{
		WorkersCount:                  cmd.opts.Concurrency,
		CancelTimeout:                 cmd.opts.Timeout,
		QuitTimeout:                   cmd.opts.QuitTimeout(),
		HelmTimeoutSeconds:            int(cmd.opts.TimeoutComponent.Seconds()),
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           cli.NewHydroformLoggerAdapter(cli.NewLogger(cmd.Verbose)),
		Profile:                       cmd.opts.Profile,
		ComponentList:                 compList,
		ResourcePath:                  resourcePath,
		InstallationResourcePath:      installResourcePath,
		Version:                       cmd.opts.Source,
		Atomic:                        cmd.opts.Atomic,
		KubeconfigSource: installConfig.KubeconfigSource{
			Path: kube.KubeconfigPath(cmd.KubeconfigPath),
		},
	}

	preinstallation(installationCfg)

	// if not verbose, use asyncui for clean output
	var callback func(deployment.ProcessUpdate)
	if !cmd.Verbose {
		ui := asyncui.AsyncUI{StepFactory: &cmd.Factory}
		callback = ui.Callback()
		if err != nil {
			return err
		}
	}

	installer, err := deployment.NewDeployment(installationCfg, overrides, callback)
	if err != nil {
		return err
	}

	return installer.StartKymaDeployment()
}

func (cmd *command) createCompList() (*installConfig.ComponentList, error) {
	var compList *installConfig.ComponentList
	if len(cmd.opts.Components) > 0 {
		compList = &installConfig.ComponentList{}
		for _, comp := range cmd.opts.Components {
			// component should be provided in the following format: componentName@namespace
			compDef := strings.Split(comp, "@")
			compName := compDef[0]
			namespace := ""
			if len(compDef) > 1 {
				namespace = compDef[1]
			}
			compList.Add(compName, namespace)
		}
	} else {
		//read component list file and marshal it to a component list entity
		compFile, err := cmd.opts.ResolveComponentsFile()
		if err != nil {
			return nil, err
		}
		compList, err = installConfig.NewComponentList(compFile)
		if err != nil {
			return nil, err
		}
	}
	return compList, nil
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
	for _, overrideList := range cmd.opts.Overrides {
		overrideList := strings.Split(overrideList, ",")
		for _, override := range overrideList {
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
	kymaVersionNames, err := cmd.installedKymaVersions()
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
		Version:        strings.Join(kymaVersionNames, ", "),
		URL:            domain.(string),
		Console:        consoleURL,
		Duration:       cmd.duration,
		Email:          string(email),
		Password:       string(pass),
	}

	return sum.Print()
}

func (cmd *command) installedKymaVersions() ([]string, error) {
	provider, err := helm.NewKymaMetadataProvider(installConfig.KubeconfigSource{
		Path: kube.KubeconfigPath(cmd.KubeconfigPath),
	})
	if err != nil {
		return nil, err
	}

	kymaVersionSet, err := provider.Versions()
	if err != nil {
		return nil, err
	}
	return kymaVersionSet.Names(), nil
}

func (cmd *command) importCertificate() error {
	ca := trust.NewCertifier(cmd.K8s)

	if !cmd.approveImportCertificate() {
		//no approval given: stop import
		ca.InstructionsAlpha()
		return nil
	}

	// get cert from cluster
	cert, err := ca.CertificateAlpha()
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

func (cmd *command) approveImportCertificate() bool {
	qImportCertsStep := cmd.NewStep("Install Kyma certificate locally")
	defer qImportCertsStep.Success()
	if cmd.avoidUserInteraction() { //do not import if user-interaction has to be avoided (suppress sudo pwd request)
		return false
	}
	return qImportCertsStep.PromptYesNo("Should the Kyma certificate be installed locally?")
}

// TODO: remove this when preinstallation API is improved: https://github.com/kyma-project/hydroform/issues/265
func preinstallation(installationCfg *installConfig.Config) {
	//Prepare cluster before Kyma installation
	preInstallerCfg := preinstaller.Config{
		InstallationResourcePath: installationCfg.InstallationResourcePath,
		Log:                      installationCfg.Log,
		KubeconfigSource:         installationCfg.KubeconfigSource,
	}

	commonRetryOpts := []retry.Option{
		retry.Delay(time.Duration(installationCfg.BackoffInitialIntervalSeconds) * time.Second),
		retry.Attempts(uint(installationCfg.BackoffMaxElapsedTimeSeconds / installationCfg.BackoffInitialIntervalSeconds)),
		retry.DelayType(retry.FixedDelay),
	}

	resourceParser := &preinstaller.GenericResourceParser{}
	resourceManager, err := preinstaller.NewDefaultResourceManager(installationCfg.KubeconfigSource, preInstallerCfg.Log, commonRetryOpts)
	if err != nil {
		log.Fatalf("Failed to create Kyma default resource manager: %v", err)
	}

	resourceApplier := preinstaller.NewGenericResourceApplier(installationCfg.Log, resourceManager)
	preInstaller, err := preinstaller.NewPreInstaller(resourceApplier, resourceParser, preInstallerCfg, commonRetryOpts)
	if err != nil {
		log.Fatalf("Failed to create Kyma pre-installer: %v", err)
	}

	result, err := preInstaller.InstallCRDs()
	if err != nil || len(result.NotInstalled) > 0 {
		log.Fatalf("Failed to install CRDs: %s", err)
	}

	result, err = preInstaller.CreateNamespaces()
	if err != nil || len(result.NotInstalled) > 0 {
		log.Fatalf("Failed to create namespaces: %s", err)
	}
}
