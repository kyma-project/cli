package module

import (
	"fmt"
	"os"

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
		Use:   "module --name=MODULE_NAME [--path=PARENT_DIR]",
		Short: "Initializes an empty module with the given name.",
		Long: `Use this command to create an empty Kyma module with the given name in the current working directory, or at some other location specified by the --path flag.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. 
With this command, you can initialize an empty module folder for the purpose of further development.

This command creates a module directory in the current working directory.
To create the module directory at a different location, use the "--path" flag.
The name of the  module directory is the same as the name of the module.
Module name must start with a letter and may only consist of alphanumeric and '.', '_', or '-' characters.
In the module directory, you'll find the template files and subdirectories corresponding to the required module structure:
    charts/       // folder containing a set of charts (each in a subfolder)
    crds/         // folder containing all CRDs required by the module
    operator/     // folder containing the operator needed to manage the module
    profiles/     // folder containing all profile settings
    channels/     // folder containing all channel settings
    config.yaml   // YAML file containing the installation configuration for any chart in the module that requires custom Helm settings
    README.md     // document explaining the module format, how it translates to OCI images, and how to develop one (can be mostly empty at the beginning)
`,

		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}

	cmd.Flags().StringVar(&o.ModuleName, "name", "", "Specifies the module name")
	cmd.Flags().StringVar(&o.ParentDir, "path", "", "Specifies the path where the module directory is created. The path must exist")
	if err := cobra.MarkFlagRequired(cmd.Flags(), "name"); err != nil {
		panic(err)
	}

	//TODO: Add validation

	return cmd
}

func (c *command) Run(args []string) error {
	var err error

	if !c.opts.NonInteractive {
		cli.AlphaWarn()
	}

	name := c.opts.ModuleName
	parentDir := c.opts.ParentDir

	if parentDir == "" {
		parentDir, err = os.Getwd()
		if err != nil {
			c.CurrentStep.Failure()
			return fmt.Errorf("Error while getting current working directory: %w", err)
		}
	}

	/* -- INIT EMPTY MODULE -- */

	c.NewStep(fmt.Sprintf("Initializing an empty module named %q in the %q parent directory", name, parentDir))
	err = scaffold.InitEmpty(osfs.New(), name, parentDir)
	if err != nil {
		c.CurrentStep.Failure()
		return err
	}
	c.CurrentStep.Success()

	return nil
}
