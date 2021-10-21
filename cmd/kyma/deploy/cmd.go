package deploy

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/model"

	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/deploy"
	"github.com/kyma-project/cli/internal/deploy/istioctl"

	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/kyma-project/cli/internal/deploy/component"
	"github.com/kyma-project/cli/internal/deploy/values"

	"github.com/kyma-project/cli/internal/resolve"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/config"
	"github.com/kyma-project/cli/internal/coredns"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/internal/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

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
	cobraCmd.Flags().IntVarP(&o.WorkerPoolSize, "concurrency", "", 4, "Set maximum number of workers to run simultaneously to deploy Kyma.")
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

	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}

	if err = cmd.setKubeClient(); err != nil {
		return err
	}

	summary := cmd.setSummary()

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	l := cli.NewLogger(o.Verbose).Sugar()
	ws, err := cmd.prepareWorkspace(l)
	if err != nil {
		return err
	}

	clusterinfo, err := clusterinfo.Discover(context.Background(), cmd.K8s.Static())
	if err != nil {
		return errors.Wrap(err, "Failed to discover underlying cluster type")
	}

	vals, err := values.Merge(cmd.opts.Sources, ws.WorkspaceDir, clusterinfo)
	if err != nil {
		return err
	}

	hasCustomDomain := cmd.opts.Domain != ""
	if _, err := coredns.Patch(l.Desugar(), cmd.K8s.Static(), hasCustomDomain, clusterinfo); err != nil {
		return err
	}

	components, err := cmd.resolveComponents(ws)
	if err != nil {
		return err
	}

	err = cmd.installPrerequisites(ws.WorkspaceDir)
	if err != nil {
		return err
	}

	err = cmd.deployKyma(l, components, vals, summary)
	if err != nil {
		return err
	}

	deployTime := time.Since(start)
	return summary.Print(deployTime)
}

func (cmd *command) deployKyma(l *zap.SugaredLogger, components component.List, vals values.Values, summary *nice.Summary) error {
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

	deployStep := cmd.NewStep("Deploying Kyma")
	deployStep.Start()

	recoResult, err := deploy.Deploy(deploy.Options{
		Components:     components,
		Values:         vals,
		StatusFunc:     cmd.printDeployStatus,
		KubeConfig:     kubeconfig,
		KymaVersion:    cmd.opts.Source,
		KymaProfile:    cmd.opts.Profile,
		Logger:         l,
		WorkerPoolSize: cmd.opts.WorkerPoolSize,
	})
	if err != nil {
		deployStep.Failuref("Failed to deploy Kyma.")
		return err
	}

	if recoResult.GetResult() == model.ClusterStatusError {
		summary.PrintFailedComponentSummary(recoResult)
		deployStep.Failure()
		return errors.Errorf("Kyma deployment failed with errors")
	}

	if recoResult.GetResult() == model.ClusterStatusReady {
		deployStep.Successf("Kyma deployed successfully!")
		return nil
	}
	return nil
}

func (cmd *command) printDeployStatus(status deploy.ComponentStatus) {
	if cmd.Verbose {
		return
	}

	switch status.State {
	case deploy.Success:
		statusStep := cmd.NewStep(fmt.Sprintf("Component '%s' deployed", status.Component))
		statusStep.Success()
	case deploy.RecoverableError:
		statusStep := cmd.NewStep(fmt.Sprintf("Component '%s' failed. Retrying...\n%s\n ", status.Component, status.Error.Error()))
		statusStep.Failure()
	case deploy.UnrecoverableError:
		statusStep := cmd.NewStep(fmt.Sprintf("Component '%s' failed, giving up.\n%s\n", status.Component, status.Error.Error()))
		statusStep.Failure()
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
		return nil, errors.Wrap(err, "Could not resolve workspace path")
	}

	wsFact, err := workspace.NewFactory(nil, wsp, l)
	if err != nil {
		return nil, errors.Wrap(err, "Could not instantiate workspace factory")
	}

	err = service.UseGlobalWorkspaceFactory(wsFact)
	if err != nil {
		return nil, errors.Wrap(err, "Could not set global workspace factory")
	}

	ws, err := wsFact.Get(cmd.opts.Source)
	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch workspace")
	}

	wsStep.Successf("Using Kyma from the workspace directory: %s", wsp)

	return ws, nil
}

func (cmd *command) resolveComponents(ws *workspace.Workspace) (component.List, error) {
	if len(cmd.opts.Components) > 0 {
		components := component.FromStrings(cmd.opts.Components)
		return components, nil
	}
	if cmd.opts.ComponentsFile != "" {
		filePath, err := resolve.File(cmd.opts.ComponentsFile, filepath.Join(ws.WorkspaceDir, "tmp"))
		if err != nil {
			return component.List{}, err
		}
		return component.FromFile(filePath)
	}

	return component.FromFile(path.Join(ws.InstallationResourceDir, "components.yaml"))
}

// avoidUserInteraction returns true if user won't provide input
func (cmd *command) avoidUserInteraction() bool {
	return cmd.NonInteractive || cmd.CI
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
		return errors.Wrap(err, "Cannot fetch Kyma version")
	}

	if currentVersion.None() {
		verifyStep.Successf("No previous Kyma version found")
		return nil
	}

	upgradeVersion, err := version.NewKymaVersion(cmd.opts.Source)
	if err != nil {
		return errors.Errorf("Version is non parsable: %s", cmd.opts.Source)
	}

	if cmd.avoidUserInteraction() {
		verifyStep.Successf("A Kyma installation with version '%s' was found. Proceeding with upgrade to '%s' in non-interactive mode. ", currentVersion.String(), upgradeVersion.String())
		return nil
	}

	upgradeScenario := currentVersion.IsCompatibleWith(upgradeVersion)
	switch upgradeScenario {
	case version.UpgradeEqualVersion:
		{
			verifyStep.Failuref("A Kyma installation was found. Current and target version are equal: %s ", currentVersion.String())
		}
	case version.UpgradeUndetermined:
		{
			verifyStep.Failuref("A Kyma installation was found, but compatibility between version '%s' and '%s' is not guaranteed. This might cause errors! ",
				currentVersion.String(), upgradeVersion.String())
		}
	case version.UpgradePossible:
		{
			verifyStep.Successf("A Kyma installation with version '%s' was found. ", currentVersion.String())
		}
	}

	if !verifyStep.PromptYesNo(fmt.Sprintf("Do you want to proceed with the upgrade to '%s'? ", upgradeVersion.String())) {
		return errors.New("Upgrade stopped by user")
	}

	return nil
}

func (cmd *command) installPrerequisites(wsp string) error {
	preReqStep := cmd.NewStep("Installing Prerequisites")

	istio, err := istioctl.New(wsp)
	if err != nil {
		return err
	}
	err = istio.Install()
	if err != nil {
		return err
	}

	preReqStep.Successf("Installed Prerequisites")
	return nil
}

func (cmd *command) setSummary() *nice.Summary {
	return &nice.Summary{
		NonInteractive: cmd.NonInteractive,
		Version:        cmd.opts.Source,
	}
}
