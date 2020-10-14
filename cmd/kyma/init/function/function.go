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
	defaultReference = "master"
	defaultBaseDir   = "/"
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
	cmd.Flags().StringVarP(&o.Runtime, "runtime", "r", defaultRuntime, `Flag used to define the environment for running your Function. Use one of these options:
	- nodejs12
	- nodejs10
	- python38`)

	// git function options
	cmd.Flags().StringVar(&o.URL, "url", "", `Git repository URL`)
	cmd.Flags().Var(&o.RepositoryName, "repository-name", `The name of the Git repository to be created`)
	cmd.Flags().StringVar(&o.Reference, "reference", defaultReference, `Commit hash or branch name`)
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", defaultBaseDir, `A directory in the repository containing the Function's sources`)

	return cmd
}

func (c *command) Run() error {
	s := c.NewStep("Generating project structure")

	if err := c.opts.setDefaults(); err != nil {
		s.Failure()
		return err
	}

	if _, err := os.Stat(c.opts.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(c.opts.Dir, 0700)
		if err != nil {
			return err
		}
	}

	configuration := workspace.Cfg{
		Runtime:   c.opts.Runtime,
		Name:      c.opts.Name,
		Namespace: c.opts.Namespace,
		Source:    c.opts.source(),
	}

	err := workspace.Initialize(configuration, c.opts.Dir)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Project generated in %s", c.opts.Dir)
	return nil
}
