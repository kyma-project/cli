package module

import (
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module/scaffold"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new Kyma CLI command
func NewCmd(o *Options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "module MODULE_NAME <PARENT_DIR>",
		Short: "Initializes an empty module with the given name in the given parent directory.",
		Long: `Use this command to create an empty Kyma module with the given name in a correspondingly named subdirectory of provided parent directory.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. 
With this command, you can initialize an empty module folder for the purpose of further development.

This command creates a directory with a given name in the target directory.
In this directory, you'll find the template files and subdirectories corresponding to the required module structure:
    charts/       // folder containing a set of charts (each in a subfolder)
    crds/         // folder containing all CRDs required by the module
    operator/     // folder containing the operator needed to manage the module
    profiles/     // folder containing all profile settings
    channels/     // folder containing all channel settings
    config.yaml   // YAML file containing installation configuration for any chart in the module that requires custom Helm settings
    README.md     // document explaining the module format, how it translates to OCI images, and how to develop one (can be mostly empty at the beginning)
`,

		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}
	cmd.Args = cobra.ExactArgs(2)
	return cmd
}

func (c *command) Run(args []string) error {
	if !c.opts.NonInteractive {
		cli.AlphaWarn()
	}

	name := args[0]
	parentDir := args[1]

	/* -- INIT EMPTY MODULE -- */

	c.NewStep(fmt.Sprintf("Initializing an empty module named %q in the %q parent directory", name, parentDir))
	err := scaffold.InitEmpty(osfs.New(), name, parentDir)
	if err != nil {
		c.CurrentStep.Failure()
		return err
	}
	c.CurrentStep.Success()

	return nil
}
