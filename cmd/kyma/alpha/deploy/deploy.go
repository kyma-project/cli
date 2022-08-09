package deploy

import (
	"errors"
	"time"

	"github.com/kyma-project/cli/internal/cli"
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
