package module

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new kyma CLI command
func NewCmd(o *Options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "module MODULE_NAME MODULE_VERSION <CONTENT_PATH> [flags]",
		Short: "Creates a module bundled as an OCI image from the contents of the given path",
		Long: `Use this command to create a Kyma module.

### Detailed description

Kyma modules are individual components that can be deplyed into a Kyma runtime. Modules are built and distributed as OCI continer images. 
This command provides the means to create such images out of a folder's contents.

This command will create a component descriptor in the descriptor path (./mod as a default) and package all the contents on the provided content path as a single layer.
Optionally, additional layers can be added with contents in other paths.

Finally, if a registry is provided, the created module will be pushed.
`,

		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}
	cmd.Args = cobra.ExactArgs(3)

	cmd.Flags().StringArrayVarP(&o.ResourcePaths, "resource", "r", []string{}, "Add an extra resource in a new layer with format <NAME:TYPE@PATH>. It is also possible to provide only a path; name will default to the last path element and type to 'helm-chart'")
	cmd.Flags().StringVar(&o.ModPath, "mod-path", "./mod", "Specifies the path where the component descriptor and module packaging will be stored. If the path already has a descriptor use the overwrite flag to overwrite it")
	cmd.Flags().StringVar(&o.PrivateKeyPath, "private-key-path", "", "Specifies the path where the private key used for signing")
	cmd.Flags().StringVar(&o.RegistryURL, "registry", "", "Repository context url for component to upload. The repository url will be automatically added to the repository contexts in the module")
	cmd.Flags().StringVarP(&o.Credentials, "credentials", "c", "", "Basic authentication credentials for the given registry in the format user:password")
	cmd.Flags().StringVarP(&o.Token, "token", "t", "", "Authentication token for the given registry (alternative to basic authentication).")
	cmd.Flags().BoolVarP(&o.Overwrite, "overwrite", "w", false, "overwrites the existing mod-path directory if it exists")
	cmd.Flags().BoolVar(&o.Insecure, "insecure", false, "Use an insecure connection to access the registry.")
	cmd.Flags().BoolVar(&o.Clean, "clean", false, "Remove the mod-path folder and all its contents at the end.")

	return cmd
}

func (c *command) Run(args []string) error {
	if !c.opts.NonInteractive {
		cli.AlphaWarn()
	}

	l := cli.NewLogger(c.opts.Verbose).Sugar()
	cfg := &module.ComponentConfig{
		Name:                 args[0],
		Version:              args[1],
		ComponentArchivePath: c.opts.ModPath,
		Overwrite:            c.opts.Overwrite,
		RegistryURL:          c.opts.RegistryURL,
	}

	/* -- CREATE ARCHIVE -- */

	// prepend the module root to the paths
	c.opts.ResourcePaths = append([]string{args[2]}, c.opts.ResourcePaths...)
	fs := osfs.New()

	c.NewStep(fmt.Sprintf("Creating module archive at %q", c.opts.ModPath))
	archive, err := module.Build(fs, cfg)
	if err != nil {
		c.CurrentStep.Failure()
		return err
	}
	c.CurrentStep.Success()

	/* -- BUNDLE RESOURCES -- */

	c.NewStep("Adding resources...")

	defs := []module.ResourceDef{}
	for _, p := range c.opts.ResourcePaths {
		rd, err := module.ResourceDefFromString(p)
		if err != nil {
			return err
		}
		defs = append(defs, rd)
	}

	if err := module.AddResources(archive, cfg, l, fs, defs...); err != nil {
		c.CurrentStep.Failure()
		return err
	}
	c.CurrentStep.Successf("Resources added")

	/* -- SIGN -- */

	c.NewStep("Signing resources...")

	c.CurrentStep.Successf("Signed component descriptor")

	/* -- PUSH -- */

	if c.opts.RegistryURL != "" {
		c.NewStep(fmt.Sprintf("Pushing image to %q", c.opts.RegistryURL))
		r := &module.Remote{
			Registry:    c.opts.RegistryURL,
			Credentials: c.opts.Credentials,
			Token:       c.opts.Token,
			Insecure:    c.opts.Insecure,
		}
		if err := module.Push(archive, r, l); err != nil {
			c.CurrentStep.Failure()
			return err
		}
		c.CurrentStep.Success()
	}

	/* -- CLEANUP -- */

	if c.opts.Clean {
		c.NewStep(fmt.Sprintf("Cleaning up mod path %q", c.opts.ModPath))
		if err := os.RemoveAll(c.opts.ModPath); err != nil {
			c.CurrentStep.Failure()
			return err
		}
		c.CurrentStep.Success()
	}

	return nil
}
