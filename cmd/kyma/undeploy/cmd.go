package undeploy

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/deploy"
	"github.com/kyma-project/cli/internal/deploy/component"
	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/kyma-project/cli/internal/version"

	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

// NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "undeploy",
		Short: "Undeploys Kyma from a running Kubernetes cluster.",
		Long:  `Use this command to undeploy Kyma from a running Kubernetes cluster.`,
		RunE:  func(cc *cobra.Command, _ []string) error { return cmd.Run(cc.Context()) },
	}

	cobraCmd.Flags().StringSliceVarP(&o.Components, "component", "", []string{}, "Provide one or more components to undeploy (e.g. --component componentName@namespace)")
	cobraCmd.Flags().StringVarP(&o.ComponentsFile, "components-file", "c", "", `Path to the components file (default "$HOME/.kyma/sources/installation/resources/components.yaml" or ".kyma-sources/installation/resources/components.yaml")`)
	cobraCmd.Flags().StringVarP(&o.WorkspacePath, "workspace", "w", "", `Path to download Kyma sources (default "$HOME/.kyma/sources" or ".kyma-sources")`)
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", "", `CAUTION: Don't use this flag except for emergency debugging. By default, the correct version is detected automatically.
	- Undeploy from a specific release, for example: "kyma undeploy --source=2.0.0"
	- Undeploy from a specific branch of the Kyma repository on kyma-project.org: "kyma undeploy --source=<my-branch-name>"
	- Undeploy from a commit (8 characters or more), for example: "kyma undeploy --source=34edf09a"
	- Undeploy from a pull request, for example "kyma undeploy --source=PR-9486"
	- Undeploy from the local sources: "kyma undeploy --source=local"`)

	cobraCmd.Flags().StringVarP(&o.Domain, "domain", "d", "", "Custom domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.Profile, "profile", "p", "",
		fmt.Sprintf("Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: %s, %s.", profileEvaluation, profileProduction))
	cobraCmd.Flags().StringVarP(&o.TLSCrtFile, "tls-crt", "", "", "TLS certificate file for the domain used for installation.")
	cobraCmd.Flags().StringVarP(&o.TLSKeyFile, "tls-key", "", "", "TLS key file for the domain used for installation.")
	cobraCmd.Flags().IntVarP(&o.WorkerPoolSize, "concurrency", "", 4, "Set maximum number of workers to run simultaneously to deploy Kyma.")
	cobraCmd.Flags().StringSliceVarP(&o.Values, "value", "", []string{}, "Set configuration values. Can specify one or more values, also as a comma-separated list (e.g. --value component.a='1' --value component.b='2' or --value component.a='1',component.b='2').")
	cobraCmd.Flags().StringSliceVarP(&o.ValueFiles, "values-file", "f", []string{}, "Path(s) to one or more JSON or YAML files with configuration values.")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 6*time.Minute, "Maximum time for the deletion")
	cobraCmd.Flags().StringVar(&o.DeleteStrategy, "delete-strategy", "system", "Specify if only Kyma resources are deleted (system) or all resources (all)")
	return cobraCmd
}

// Run runs the command
func (cmd *command) Run(ctx context.Context) error {
	var err error

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}

	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Cannot initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	if !cmd.opts.NonInteractive {
		if err := cli.DetectManagedEnvironment(ctx, cmd.K8s, cmd.Factory.NewStep("")); err != nil {
			return err
		}
	}

	kubeconfigPath := kube.KubeconfigPath(cmd.KubeconfigPath)
	kubeconfig, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "failed to read kubeconfig")
	}

	// Prepare workspace
	if err := cmd.resolveVersion(); err != nil {
		return err
	}

	wsStep := cmd.NewStep(fmt.Sprintf("Fetching Kyma sources (%s)", cmd.opts.Source))
	l := cli.NewLogger(cmd.opts.Verbose).Sugar()
	ws, err := deploy.PrepareWorkspace(cmd.opts.WorkspacePath, cmd.opts.Source, wsStep, !cmd.avoidUserInteraction(), cmd.opts.IsLocal(), l)
	if err != nil {
		return err
	}

	clusterInfo, err := clusterinfo.Discover(ctx, cmd.K8s.Static())
	if err != nil {
		return errors.Wrap(err, "failed to discover underlying cluster type")
	}

	vals, err := values.Merge(cmd.opts.Sources, ws.WorkspaceDir, clusterInfo)
	if err != nil {
		return err
	}

	components, err := component.Resolve(cmd.opts.Components, cmd.opts.ComponentsFile, ws)
	if err != nil {
		return err
	}

	undeployStep := cmd.NewStep("Undeploying Kyma")
	undeployStep.Start()

	// do the undeploy
	recoResult, err := deploy.Undeploy(deploy.Options{
		Components:     components,
		Values:         vals,
		StatusFunc:     cmd.printDeployStatus,
		KubeConfig:     kubeconfig,
		KymaVersion:    cmd.opts.Source,
		KymaProfile:    cmd.opts.Profile,
		Logger:         l,
		WorkerPoolSize: cmd.opts.WorkerPoolSize,
		DeleteStrategy: cmd.opts.DeleteStrategy,
	})
	if err != nil {
		undeployStep.Failuref("Failed to undeploy Kyma.")
		return err
	}

	if recoResult.GetResult() == model.ClusterStatusDeleteError {
		undeployStep.Failure()
		cmd.setSummary().PrintFailedComponentSummary(recoResult)
		return errors.Errorf("kyma undeployment failed")
	}

	if recoResult.GetResult() == model.ClusterStatusDeleted {
		undeployStep.Success()
	}
	undeployStep.Successf("Kyma undeployed successfully!")
	return nil
}

func (cmd *command) printDeployStatus(status deploy.ComponentStatus) {
	if cmd.Verbose {
		return
	}

	switch status.State {
	case deploy.Success:
		statusStep := cmd.NewStep(fmt.Sprintf("Component '%s' deleted", status.Component))
		statusStep.Success()
	case deploy.RecoverableError:
		if deploy.PrintedStatus[status.Component] {
			return
		}
		deploy.PrintedStatus[status.Component] = true
		statusStep := cmd.NewStep(fmt.Sprintf("Component '%s' failed. Retrying...\n%s\n ", status.Component, status.Error.Error()))
		statusStep.Failure()
	case deploy.UnrecoverableError:
		statusStep := cmd.NewStep(fmt.Sprintf("Component '%s' failed \n%s\n", status.Component, status.Error.Error()))
		statusStep.Failure()
	}
}

// avoidUserInteraction returns true if user won't provide input
func (cmd *command) avoidUserInteraction() bool {
	return cmd.NonInteractive || cmd.CI
}

func (cmd *command) setSummary() *nice.Summary {
	return &nice.Summary{
		NonInteractive: cmd.NonInteractive,
		Version:        cmd.opts.Source,
	}
}

// resolveVersion determines which version of kyma has to be used for undeploy.
// Version is automatically detected from the cluster and can be optionally overridden via flag.
func (cmd *command) resolveVersion() error {
	if cmd.opts.Source != "" {
		return nil
	}
	v, err := version.GetCurrentKymaVersion(cmd.K8s)
	if err != nil {
		return errors.Wrap(err, "could not determine Kyma version to undeploy")
	}

	switch {
	case v.None():
		return errors.New("could not determine Kyma version to undeploy: version info empty")
	case v.IsKyma1():
		return errors.Errorf("Kyma version [%s] not supported", v.String())
	}

	cmd.opts.Source = v.String()
	return nil
}
