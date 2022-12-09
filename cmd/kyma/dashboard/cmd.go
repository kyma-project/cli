package dashboard

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/dashboard"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

// NewCmd creates a new dashboard command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Runs the Kyma dashboard locally and opens it directly in a web browser.",
		Long:  `Use this command to run the Kyma dashboard locally in a docker container and open it directly in a web browser. This command only works with a local installation of Kyma.`,
		RunE:  func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Port, "port", "p", "3001", `Specify the port on which the local dashboard will be exposed.`)
	cmd.Flags().StringVar(&o.ContainerName, "container-name", "kyma-dashboard", `Specify the name of the local container.`)

	return cmd
}

// Run runs the command
func (cmd *command) Run() error {
	var err error

	if err = cmd.opts.validateFlags(); err != nil {
		return err
	}
	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	cmd.NewStep(fmt.Sprintf("Starting dashboard container: %s:", cmd.opts.ContainerName))
	dash := dashboard.New(cmd.opts.ContainerName, cmd.opts.Port, cmd.KubeconfigPath, cmd.Verbose)

	if err := dash.Start(); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}

	// make sure the dahboard container always stops at the end
	cmd.Finalizers.Add(dash.StopFunc(context.Background(), func(i ...interface{}) { fmt.Print(i...) }))

	cmd.CurrentStep.Successf("Started container: %s", cmd.opts.ContainerName)

	cmd.NewStep("Opening the dashboard on the browser")
	if err := dash.Open(""); err != nil {
		cmd.CurrentStep.Failure()
		return err
	}
	cmd.CurrentStep.Success()

	return dash.Watch(context.Background())
}
