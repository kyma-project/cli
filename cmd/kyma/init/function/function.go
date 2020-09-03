package function

import (
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
	"os"
)

const (
	defaultRuntime   = "nodejs12"
	defaultNamespace = "default"
	defaultName      = "first-function"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new init command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		opts:    o,
		Command: cli.Command{Options: o.Options},
	}
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Creates local resources for your Function.",
		Long: `Use this command to create the local workspace with the default structure of your Function's code and dependencies. Update this configuration to your references and apply it to a Kyma cluster. 
Use the flags to specify the initial configuration for your Function or to choose the location for your project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Run()
		},
	}

	cmd.Flags().StringVar(&o.Name, "name", defaultName, `Function name.`)
	cmd.Flags().StringVar(&o.Namespace, "namespace", defaultNamespace, `Namespace to which you want to apply your Function.`)
	cmd.Flags().StringVarP(&o.Dir, "dir", "d", "", `Full path to the directory where you want to save the project.`)
	cmd.Flags().StringVarP(&o.Runtime, "runtime", "r", defaultRuntime, `Flag used to define the environment for running you Function. Use one of:
	- nodejs12
	- nodejs10
	- python38`)

	return cmd
}

func (c *command) Run() error {
	s := c.NewStep("Generating project structure")
	if c.opts.Dir == "" {
		var err error
		c.opts.Dir, err = os.Getwd()
		if err != nil {
			s.Failure()
			return err
		}
	}

	configuration := workspace.Cfg{
		Runtime:   c.opts.Runtime,
		Name:      c.opts.Name,
		Namespace: c.opts.Namespace,
		Source: workspace.SourceInline{
			BaseDir: c.opts.Dir,
		},
	}

	err := workspace.Initialize(configuration, c.opts.Dir)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Project generated in %s", c.opts.Dir)
	return nil
}
