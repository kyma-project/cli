package deploy

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/logger"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-incubator/reconciler/pkg/scheduler"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/component"
	"github.com/kyma-project/cli/internal/coredns"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/k3d"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/trust"
	"github.com/kyma-project/cli/internal/version"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	//Register all reconcilers
	_ "github.com/kyma-incubator/reconciler/pkg/reconciler/instances"
)

type command struct {
	cli.Command
	opts *Options
}

//NewCmd creates a new deploy command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "deploy",
		Short:   "Deploys Kyma on a running Kubernetes cluster.",
		Long:    "Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.",
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run(cmd.opts) },
		Aliases: []string{"d"},
	}
	cobraCmd.Flags().StringSliceVarP(&o.Components, "component", "", []string{}, "Provide one or more components to deploy (e.g. --component componentName@namespace)")
	cobraCmd.Flags().StringVarP(&o.ComponentsFile, "components-file", "c", "", `Path to the components file (default "$HOME/.kyma/sources/installation/resources/components.yaml" or ".kyma-sources/installation/resources/components.yaml")`)
	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", "", `Path to download Kyma sources (default "$HOME/.kyma/sources" or ".kyma-sources")`)
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", defaultKymaVersion, `Installation source:
	- Deploy a specific release, for example: "kyma deploy --source=2.0.0"
	- Deploy a specific branch of the Kyma repository on kyma-project.org: "kyma deploy --source=<my-branch-name>"
	- Deploy a commit (8 characters or more), for example: "kyma deploy --source=34edf09a"
	- Deploy a pull request, for example "kyma deploy --source=PR-9486"
	- Deploy the local sources: "kyma deploy --source=local"`)

	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", "", "Custom domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "",
		fmt.Sprintf("Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: %s, %s.", profileEvaluation, profileProduction))
	cobraCmd.Flags().StringVarP(&o.TLSCrtFile, "tls-crt", "", "", "TLS certificate file for the domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSKeyFile, "tls-key", "", "", "TLS key file for the domain used for installation.")
	cobraCmd.Flags().StringSliceVarP(&o.Values, "value", "", []string{}, "Set configuration values. Can specify one or more values, also as a comma-separated list (e.g. --value component.a='1' --value component.b='2' or --value component.a='1',component.b='2').")
	cobraCmd.Flags().StringSliceVarP(&o.ValueFiles, "values-file", "f", []string{}, "Path(s) to one or more JSON or YAML files with configuration values.")

	return cobraCmd
}

func (cmd *command) Run(o *Options) error {
	var err error

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if err = cmd.setKubeClient(); err != nil {
		return err
	}

	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	l := logger.NewLogger(o.Verbose)
	ws, err := cmd.workspaceBuilder(l)
	if err != nil {
		return err
	}

	values, err := mergeValues(cmd.opts, ws, cmd.K8s.Static())
	if err != nil {
		return err
	}

	isK3d, err := k3d.IsK3dCluster(cmd.K8s.Static())
	if err != nil {
		return err
	}

	hasCustomDomain := cmd.opts.Domain != ""
	if _, err := coredns.Patch(zap.NewNop(), cmd.K8s.Static(), hasCustomDomain, isK3d); err != nil {
		return err
	}

	components, err := cmd.createCompListWithOverrides(ws, values)
	if err != nil {
		return err
	}

	err = cmd.deployKyma(components)
	if err != nil {
		return err
	}

	if err := cmd.importCertificate(); err != nil {
		return err
	}

	// TODO: print summary after deploy

	return nil
}

func (cmd *command) deployKyma(comps component.List) error {
	kubeconfigPath := kube.KubeconfigPath(cmd.KubeconfigPath)
	kubeconfig, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Could not read kubeconfig")
	}

	localScheduler := scheduler.NewLocalScheduler(
		scheduler.WithPrerequisites(cmd.buildCompList(comps.Prerequisites)...),
		scheduler.WithStatusFunc(cmd.printDeployStatus))

	componentsToInstall := append(comps.Prerequisites, comps.Components...)
	err = localScheduler.Run(context.TODO(), &keb.Cluster{
		Kubeconfig: string(kubeconfig),
		KymaConfig: keb.KymaConfig{
			Version:    cmd.opts.Source,
			Profile:    cmd.opts.Profile,
			Components: componentsToInstall,
		},
	})
	if err != nil {
		return errors.Wrap(err, "Failed to deploy Kyma")
	}
	return nil
}

func (cmd *command) workspaceBuilder(l *zap.SugaredLogger) (*workspace.Workspace, error) {
	if err := cmd.decideVersionUpgrade(); err != nil {
		return nil, err
	}
	wsStep := cmd.NewStep(fmt.Sprintf("Fetching Kyma (%s)", cmd.opts.Source))

	//Check if workspace is empty or not
	if cmd.opts.Source != VersionLocal {
		_, err := os.Stat(cmd.opts.WorkspacePath)
		// workspace already exists
		if !os.IsNotExist(err) && !cmd.avoidUserInteraction() {
			isWorkspaceEmpty, err := files.IsDirEmpty(cmd.opts.WorkspacePath)
			if err != nil {
				return nil, err
			}
			// if workspace used is not the default one and it is not empty,
			// then ask for permission to delete its existing files
			if !isWorkspaceEmpty && cmd.opts.WorkspacePath != getDefaultWorkspacePath() {
				if !wsStep.PromptYesNo(fmt.Sprintf("Existing files in workspace folder '%s' will be deleted. Are you sure you want to continue? ", cmd.opts.WorkspacePath)) {
					wsStep.Failure()
					return nil, fmt.Errorf("aborting deployment")
				}
			}
		}
	}

	wsp, err := cmd.opts.ResolveLocalWorkspacePath()
	if err != nil {
		return &workspace.Workspace{}, err
	}

	wsFact, err := workspace.NewFactory(nil, wsp, l)
	if err != nil {
		return &workspace.Workspace{}, err
	}
	err = service.UseGlobalWorkspaceFactory(wsFact)
	if err != nil {
		return nil, err
	}

	ws, err := wsFact.Get(cmd.opts.Source)
	if err != nil {
		return &workspace.Workspace{}, err
	}

	wsStep.Successf("Fetching kyma from workspace folder: %s", wsp)

	return ws, nil
}

func (cmd *command) buildCompList(comps []keb.Component) []string {
	var compSlice []string
	for _, c := range comps {
		compSlice = append(compSlice, c.Component)
	}
	return compSlice
}

func (cmd *command) createCompListWithOverrides(ws *workspace.Workspace, overrides map[string]interface{}) (component.List, error) {
	var compList component.List
	if len(cmd.opts.Components) > 0 {
		compList = component.FromStrings(cmd.opts.Components, overrides)
		return compList, nil
	}
	if cmd.opts.ComponentsFile != "" {
		return component.FromFile(ws, cmd.opts.ComponentsFile, overrides)
	}
	compFile := path.Join(ws.InstallationResourceDir, "components.yaml")
	return component.FromFile(ws, compFile, overrides)
}

func (cmd *command) printDeployStatus(component string, msg *reconciler.CallbackMessage) {
	fmt.Printf("Component %s has status %s\n", component, msg.Status)
}

// avoidUserInteraction returns true if user won't provide input
func (cmd *command) avoidUserInteraction() bool {
	return cmd.NonInteractive || cmd.CI
}

func (cmd *command) importCertificate() error {
	ca := trust.NewCertifier(cmd.K8s)

	if !cmd.approveImportCertificate() {
		//no approval given: stop import
		ca.InstructionsKyma2()
		return nil
	}

	// get cert from cluster
	cert, err := ca.CertificateKyma2()
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

func (cmd *command) setKubeClient() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Cannot initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}
	return nil
}

func (cmd *command) decideVersionUpgrade() error {
	verifyStep := cmd.NewStep("Verifying Kyma version compatibility")

	currentVersion, err := version.GetCurrentKymaVersion(cmd.K8s)
	if err != nil {
		return errors.Wrap(err, "Cannot fetch kyma version")
	}

	if currentVersion.HasNoVersion() {
		verifyStep.Successf("No previous Kyma version found")
		return nil
	}

	if currentVersion.IsKyma1() {
		if cmd.avoidUserInteraction() {
			verifyStep.Failuref("A kyma v1 installation (%s) was found. Please use interactive mode to confirm the upgrade", currentVersion.String())
		}
		if !verifyStep.PromptYesNo(fmt.Sprintf("A kyma v1 installation (%s) was found. Do you want to proceed with the upgrade (%s)? ", currentVersion.String(), cmd.opts.Source)) {
			return errors.New("Upgrade stopped by user")
		}
	}

	if currentVersion.IsKyma2() {
		if !verifyStep.PromptYesNo(fmt.Sprintf("A kyma v2 installation (%s) was found. Do you want to proceed with the upgrade? ", currentVersion.String())) {
			return errors.New("Upgrade stopped by user")
		}
	}

	if !currentVersion.IsReleasedVersion() {
		// Assume we are upgrading from PR-XXX or main or branch
		if !verifyStep.PromptYesNo(fmt.Sprintf("A kyma installation with version (%s) was found. Do you want to proceed with the upgrade? ", currentVersion.String())) {
			return errors.New("Upgrade stopped by user")
		}
	}

	upgradeVersion, err := version.NewKymaVersion(cmd.opts.Source)
	if err != nil {
		return errors.Errorf("Version is non parsable: %s", cmd.opts.Source)
	}

	upgradeDecision := currentVersion.IsCompatibleWith(upgradeVersion)
	switch upgradeDecision {
	case version.UpgradeEqualVersion:
		{
			verifyStep.Failuref("Current and next Kyma version are equal: %s", currentVersion.String())
		}
	case version.UpgradeUndetermined:
		{
			verifyStep.Failuref("Cannot check compatibility between version '%s' and '%s'. This might cause errors!",
				currentVersion.String(), upgradeVersion.String())
		}
	case version.UpgradePossible:
		{
			verifyStep.Success()
		}
	}
	//seemless upgrade unnecessary or cannot be warrantied, asking user for approval
	incompatibleStep := cmd.NewStep("Continue Kyma upgrade")
	if cmd.avoidUserInteraction() || incompatibleStep.PromptYesNo("Do you want to proceed with the upgrade? ") {
		incompatibleStep.Success()
		return nil
	}
	incompatibleStep.Failure()
	return fmt.Errorf("upgrade stopped by user")
}
