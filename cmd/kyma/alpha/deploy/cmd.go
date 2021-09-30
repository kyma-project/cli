package deploy

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/kyma-incubator/reconciler/pkg/keb"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-incubator/reconciler/pkg/scheduler"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/component"
	"github.com/kyma-project/cli/internal/config"
	"github.com/kyma-project/cli/internal/coredns"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/istio"
	"github.com/kyma-project/cli/internal/k3d"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/internal/trust"
	"github.com/kyma-project/cli/internal/version"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	//Register all reconcilers
	_ "github.com/kyma-incubator/reconciler/pkg/reconciler/instances"
)

const (
	dashboardURL = "https://dashboard.kyma.cloud.sap"
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
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", config.DefaultKyma2Version, `Installation source:
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
	start := time.Now()

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

	l := cli.NewLogger(o.Verbose).Sugar()
	ws, err := cmd.prepareWorkspace(l)
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
	if _, err := coredns.Patch(l.Desugar(), cmd.K8s.Static(), hasCustomDomain, isK3d); err != nil {
		return err
	}

	components, err := cmd.createComponentsWithOverrides(ws, values)
	if err != nil {
		return err
	}

	err = cmd.installPrerequisites(ws.WorkspaceDir)
	if err != nil {
		return err
	}

	err = cmd.deployKyma(l, components)
	if err != nil {
		return err
	}

	if err := cmd.importCertificate(); err != nil {
		return err
	}

	return cmd.printSummary(values, time.Since(start))
}

func (cmd *command) deployKyma(l *zap.SugaredLogger, components component.List) error {
	kubeconfigPath := kube.KubeconfigPath(cmd.KubeconfigPath)
	kubeconfig, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Could not read kubeconfig")
	}

	undo := zap.RedirectStdLog(l.Desugar())
	defer undo()

	if !cmd.opts.Verbose {
		stderr := os.Stderr
		os.Stderr = nil
		defer func() { os.Stderr = stderr }()
	}

	localScheduler := scheduler.NewLocalScheduler(
		scheduler.WithLogger(l),
		scheduler.WithPrerequisites(cmd.componentNames(components.Prerequisites)...),
		scheduler.WithStatusFunc(cmd.printDeployStatus))

	componentsToInstall := append(components.Prerequisites, components.Components...)
	step := cmd.NewStep("Deploying Kyma")
	step.Start()

	err = localScheduler.Run(context.TODO(), &keb.Cluster{
		Kubeconfig: string(kubeconfig),
		KymaConfig: keb.KymaConfig{
			Version:    cmd.opts.Source,
			Profile:    cmd.opts.Profile,
			Components: componentsToInstall,
		},
	})
	if err != nil {
		step.Failuref("Failed to deploy Kyma.")
		return err
	}

	step.Successf("Kyma deployed successfully!")
	return nil
}

func (cmd *command) printDeployStatus(component string, msg *reconciler.CallbackMessage) {
	if cmd.Verbose {
		return
	}

	switch msg.Status {
	case reconciler.StatusSuccess:
		step := cmd.NewStep(fmt.Sprintf("Component '%s' deployed", component))
		step.Success()
	case reconciler.StatusFailed:
		step := cmd.NewStep(fmt.Sprintf("Component '%s' failed. Retrying...", component))
		step.Failure()
	case reconciler.StatusError:
		step := cmd.NewStep(fmt.Sprintf("Component '%s' failed and terminated", component))
		step.Failure()
	}
}

func (cmd *command) prepareWorkspace(l *zap.SugaredLogger) (*workspace.Workspace, error) {
	if err := cmd.decideVersionUpgrade(); err != nil {
		return nil, err
	}
	wsStep := cmd.NewStep(fmt.Sprintf("Fetching Kyma sources(%s)", cmd.opts.Source))

	if cmd.opts.Source != VersionLocal {
		_, err := os.Stat(cmd.opts.WorkspacePath)
		if !os.IsNotExist(err) && !cmd.avoidUserInteraction() {
			isWorkspaceEmpty, err := files.IsDirEmpty(cmd.opts.WorkspacePath)
			if err != nil {
				return nil, err
			}
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

	wsStep.Successf("Using Kyma from the workspace directory: %s", wsp)

	return ws, nil
}

func (cmd *command) componentNames(comps []keb.Component) []string {
	var names []string
	for _, c := range comps {
		names = append(names, c.Component)
	}
	return names
}

func (cmd *command) createComponentsWithOverrides(ws *workspace.Workspace, overrides map[string]interface{}) (component.List, error) {
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
	return qImportCertsStep.PromptYesNo("Do you want to install the Kyma certificate locally?")
}

func (cmd *command) printSummary(overrides map[string]interface{}, duration time.Duration) error {
	domain, ok := overrides["global.domainName"]
	if !ok {
		return errors.New("domain not found in overrides")
	}

	sum := nice.Summary{
		NonInteractive: cmd.NonInteractive,
		Version:        cmd.opts.Source,
		URL:            domain.(string),
		Dashboard:      dashboardURL,
		Duration:       duration,
	}

	return sum.Print()
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

	if currentVersion.None() {
		verifyStep.Successf("No previous Kyma version found")
		return nil
	}

	if !currentVersion.IsReleasedVersion() {
		// Assume we are upgrading from PR-XXX or main or branch
		if !verifyStep.PromptYesNo(fmt.Sprintf("A kyma installation with version (%s) was found. Do you want to proceed with the upgrade (%s)? ", currentVersion.String(), cmd.opts.Source)) {
			return errors.New("Upgrade stopped by user")
		}
	} else if currentVersion.IsKyma1() {
		if cmd.avoidUserInteraction() {
			verifyStep.Failuref("A kyma v1 installation (%s) was found. Please use interactive mode to confirm the upgrade", currentVersion.String())
		}
		if !verifyStep.PromptYesNo(fmt.Sprintf("A kyma v1 installation (%s) was found. Do you want to proceed with the upgrade (%s)? ", currentVersion.String(), cmd.opts.Source)) {
			return errors.New("Upgrade stopped by user")
		}
	} else if currentVersion.IsKyma2() {
		if !verifyStep.PromptYesNo(fmt.Sprintf("A kyma v2 installation (%s) was found. Do you want to proceed with the upgrade? ", currentVersion.String())) {
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

func (cmd *command) installPrerequisites(wsp string) error {
	preReqStep := cmd.NewStep("Installing Prerequisites")

	istioctl, err := istio.New(wsp)
	if err != nil {
		return err
	}
	err = istioctl.Install()
	if err != nil {
		return err
	}

	preReqStep.Successf("Installed Prerequisites")
	return nil
}
