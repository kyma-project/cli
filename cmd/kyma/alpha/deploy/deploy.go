package deploy

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/spf13/cobra"
)

type command struct {
	cli.Command
	opts *Options
}

//NewCmd creates a new Kyma CLI command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys Kyma on a running Kubernetes cluster.",
		Long: `Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.
`,
		RunE: func(_ *cobra.Command, _ []string) error { return cmd.RunWithTimeout() },
	}

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

	if cmd.opts.DryRun {
		return cmd.dryRun()
	}

	var err error
	if err = cmd.setKubeClient(); err != nil {
		return err
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "failed to initialize the Kubernetes client from given kubeconfig")
	}

	/*
		if err := cmd.decideVersionUpgrade(); err != nil {
			return err
		}
	*/

	return cmd.deploy(start)
}

func (cmd *command) dryRun() error {
	return nil
}

func (cmd *command) deploy(start time.Time) error {
	//TODO: Implement
	return nil
}

func (cmd *command) setKubeClient() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "failed to initialize the Kubernetes client from given kubeconfig")
	}
	return nil
}
