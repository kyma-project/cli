package deploy

import (
	"log"
	"os"
	"time"

	"github.com/kyma-project/cli/internal/config"
	"github.com/kyma-project/cli/internal/deploy"
	"github.com/kyma-project/cli/internal/kustomize"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
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
	cobraCmd.Flags().StringVarP(&o.Source, "source", "s", config.DefaultManagerVersion, `Installation source:
	- Deploy a specific release  of the lifecycle and module manager: "kyma deploy --source=2.0.0"
	- Deploy a specific branch of the lifecycle and module manager: "kyma deploy --source=<my-branch-name>"
	- Deploy a commit (8 characters or more) of the lifecycle and module manager: "kyma deploy --source=34edf09a"
	- Deploy a pull request, for example "kyma deploy --source=PR-9486"
	- Deploy the local sources  of the lifecycle and module manager: "kyma deploy --source=local"`)
	cobraCmd.Flags().StringArrayVarP(&o.Modules, "module", "m", []string{}, `Provide one or more modules to activate after the deployment is finished, for example:
	- With short-hand notation: "--module name@namespace"
	- With verbose JSON structure "--module '{"name": "componentName","namespace": "componenNamespace","url": "componentUrl","version": "1.2.3"}'`)
	cobraCmd.Flags().StringVarP(&o.ModulesFile, "modules-file", "f", "", `Path to file containing a list of modules.`)
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
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "failed to initialize the Kubernetes client from given kubeconfig")
	}

	return cmd.deploy(start)
}

func (cmd *command) deploy(start time.Time) error {
	if cmd.opts.DryRun {
		return cmd.dryRun()
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()

	if err := cmd.initialSetup(); err != nil {
		return err
	}

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

	if err := deploy.Operators(cmd.opts.Source, cmd.K8s, false); err != nil {
		return err
	}
	var err error

	if err != nil {
		deployStep.Failuref("Failed to deploy Kyma.")
		return err
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
	return deploy.Operators(cmd.opts.Source, cmd.K8s, true)
}
