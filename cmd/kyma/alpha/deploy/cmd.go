package deploy

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kyma-project/cli/internal/deploy"
	"github.com/kyma-project/cli/internal/kustomize"
	"github.com/kyma-project/cli/pkg/errs"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type command struct {
	cli.Command
	opts *Options
}

// NewCmd creates a new deploy command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "deploy",
		Short:   "Deploys Kyma on a running Kubernetes cluster.",
		Long:    "Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.",
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.RunWithTimeout() },
		Aliases: []string{"d"},
	}
	cobraCmd.Flags().StringArrayVarP(&o.Kustomizations, "kustomization", "k", []string{}, `Provide one or more kustomizations to deploy. Each occurrence of the flag accepts a URL with an optional reference (commit, branch, or release) in the format URL@ref or a local path to the directory of the kustomization file.
	Defaults to deploying Lifecycle Manager and Module Manager from GitHub main branch.
	Examples:
	- Deploy a specific release of the Lifecycle Manager: "kyma deploy -k https://github.com/kyma-project/lifecycle-manager/operator/config/default@1.2.3"
	- Deploy a local Module Manager: "kyma deploy --kustomization /path/to/repo/module-manager/operator/config/default"
	- Deploy a branch of Lifecycle Manager with a custom URL: "kyma deploy -k https://gitlab.com/forked-from-github/lifecycle-manager/operator/config/default@feature-branch-1"
	- Deploy the main branch of Lifecycle Manager while using local sources of Module Manager: "kyma deploy -k /path/to/repo/module-manager/operator/config/default -k https://github.com/kyma-project/lifecycle-manager/operator/config/default@main"`)
	cobraCmd.Flags().StringArrayVarP(&o.Modules, "module", "m", []string{}, `Provide one or more modules to activate after the deployment is finished. Example: "--module name@namespace" (namespace is optional).`)
	cobraCmd.Flags().StringVarP(&o.ModulesFile, "modules-file", "f", "", `Path to file containing a list of modules.`)
	cobraCmd.Flags().StringVarP(&o.Channel, "channel", "c", "stable", `Select which channel to deploy from: stable, fast, nightly.`)
	cobraCmd.Flags().StringVar(&o.Channel, "kyma-cr", "", `Provide a custom Kyma CR file for the deployment.`)

	// TODO remove this flag when module templates can be fetched from release.
	// Might be worth keeping this flag with another name to install extra templates??
	cobraCmd.Flags().StringArrayVar(&o.Templates, "template", []string{}, `Provide one or more module templates to deploy.
	WARNING: This is a temporary flag for development and will be removed soon.`)

	cobraCmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "Renders the Kubernetes manifests without actually applying them.")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "t", 20*time.Minute, "Maximum time for the deployment.")

	return cobraCmd
}

func (cmd *command) RunWithTimeout() error {
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if err := cmd.opts.validateFlags(); err != nil {
		return err
	}

	timeout := time.After(cmd.opts.Timeout)
	errChan := make(chan error)
	go func() {
		errChan <- cmd.run()
	}()

	for {
		select {
		case <-timeout:
			msg := "Timeout reached while waiting for deployment to complete"
			timeoutStep := cmd.NewStep(msg)
			timeoutStep.Failure()
			return errors.New(msg)
		case err := <-errChan:
			return err
		}
	}
}

func (cmd *command) run() error {
	start := time.Now()

	var err error
	if cmd.K8s, err = kube.NewFromConfigWithTimeout("", cmd.KubeconfigPath, cmd.opts.Timeout); err != nil {
		return errors.Wrap(err, "failed to initialize the Kubernetes client from given kubeconfig")
	}

	return cmd.deploy(start)
}

func (cmd *command) deploy(start time.Time) error {
	if err := cmd.initialSetup(); err != nil {
		return err
	}

	if cmd.opts.DryRun {
		return cmd.dryRun()
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()

	summary := &nice.Summary{
		NonInteractive: cmd.NonInteractive,
		Version:        "",
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

	hasKyma, err := deploy.Bootstrap(cmd.opts.Kustomizations, cmd.K8s, false)
	if err != nil {
		deployStep.Failuref("Failed to deploy Kyma.")
		return err
	}

	// wait for operators to be ready
	if err := cmd.waitForOperators(); err != nil {
		return err
	}

	// deploy modules and kyma CR
	if hasKyma {
		// TODO change to fetch templates from release artifacts
		modStep := cmd.NewStep("Modules deployed")
		for _, t := range cmd.opts.Templates {
			b, err := os.ReadFile(t)
			if err != nil {
				modStep.Failuref("Failed to deploy modules")
				return err
			}
			if err := cmd.K8s.Apply(b); err != nil {
				modStep.Failuref("Failed to deploy modules")
				return err
			}
		}
		modStep.Success()

		kymaStep := cmd.NewStep("Kyma CR deployed")
		if err := deploy.Kyma(cmd.K8s, cmd.opts.Channel, cmd.opts.KymaCR, false); err != nil {
			kymaStep.Failuref("Failed to deploy Kyma CR")
			return err
		}
		kymaStep.Success()

	} else {
		deployStep.LogInfo("There was no Kyma CRD present in the prerequisites, no modules will be installed.")
	}

	deployStep.Successf("Kyma deployed successfully!")

	deployTime := time.Since(start)
	return summary.Print(deployTime)
}

func (cmd *command) initialSetup() error {
	s := cmd.NewStep("Setting up kustomize...")
	if err := kustomize.Setup(s, true); err != nil {
		log.Fatal(err)
	}
	s.Successf("Kustomize ready")
	return nil
}

func (cmd *command) dryRun() error {
	hasKyma, err := deploy.Bootstrap(cmd.opts.Kustomizations, cmd.K8s, true)
	if err != nil {
		return err
	}

	if hasKyma {
		// TODO change to fetch templates from release artifacts
		for _, t := range cmd.opts.Templates {
			b, err := os.ReadFile(t)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n---\n", string(b))
		}

		if err := deploy.Kyma(cmd.K8s, cmd.opts.Channel, cmd.opts.KymaCR, true); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *command) waitForOperators() error {
	errChan := make(chan error)
	defer close(errChan)

	go func() {
		err := cmd.K8s.WaitDeploymentStatus("kcp-system", "lifecycle-manager-controller-manager", appsv1.DeploymentAvailable, corev1.ConditionTrue)
		lifecycleStep := cmd.NewStep("Lifecycle Manager deployed")
		if err != nil {
			lifecycleStep.Failuref("Failed to deploy Lifecycle Manager")
		} else {
			lifecycleStep.Success()
		}
		errChan <- err
	}()

	go func() {
		err := cmd.K8s.WaitDeploymentStatus("kcp-system", "module-manager-controller-manager", appsv1.DeploymentAvailable, corev1.ConditionTrue)
		moduleStep := cmd.NewStep("Module Manager deployed")
		if err != nil {
			moduleStep.Failuref("Failed to deploy Module Manager")
		} else {
			moduleStep.Success()
		}
		errChan <- err
	}()

	// Merge errors from all async calls (2)
	return errs.MergeErrors(<-errChan, <-errChan)
}
