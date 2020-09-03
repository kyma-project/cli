package init

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/resources/types"
	"github.com/kyma-project/cli/internal/workspace"
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
		Use:   "init",
		Short: "Init something",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run()
		},
	}

	cmd.Flags().StringVarP(&o.Dir, "dir", "d", "", `Location of the generated project.`)
	cmd.Flags().StringVarP(&o.Name, "name", "n", defaultName, `Name of the function`)
	cmd.Flags().StringVar(&o.Namespace, "namespace", defaultNamespace, `Namespace of the function`)
	cmd.Flags().StringVarP(&o.Runtime, "runtime", "r", defaultRuntime, `Flag used for determinate function runtime. Use one of:
	- nodejs12
	- nodejs10
	- python38`)

	return cmd
}

func (c *command) run() error {
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
		Runtime:    types.Runtime(c.opts.Runtime),
		Name:       c.opts.Name,
		Namespace:  c.opts.Namespace,
		SourcePath: c.opts.Dir,
	}

	err := workspace.Initialize(configuration, c.opts.Dir)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Project generated in %s", c.opts.Dir)
	return nil
}
