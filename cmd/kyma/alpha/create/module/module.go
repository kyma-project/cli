package module

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module"
	"github.com/kyma-project/cli/pkg/module/oci"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

// NewCmd creates a new Kyma CLI command
func NewCmd(o *Options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "module OCI_IMAGE_NAME MODULE_VERSION <CONTENT_PATH> [flags]",
		Short: "Creates a module bundled as an OCI image with the given OCI image name from the contents of the given path",
		Long: `Use this command to create a Kyma module and bundle it as an OCI image.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. Modules are built and distributed as OCI container images. 
With this command, you can create such images out of a folder's contents.

This command creates a component descriptor in the descriptor path (./mod as a default) and packages all the contents on the provided content path as a single layer.
Optionally, you can create additional layers with contents in other paths.

Finally, if a registry is provided, the created module is pushed.
`,

		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}
	cmd.Args = cobra.ExactArgs(3)

	cmd.Flags().StringArrayVarP(&o.ResourcePaths, "resource", "r", []string{}, "Add an extra resource in a new layer with format <NAME:TYPE@PATH>. It is also possible to provide only a path; name will default to the last path element and type to 'helm-chart'")
	cmd.Flags().StringVar(&o.ModPath, "mod-path", "./mod", "Specifies the path where the component descriptor and module packaging will be stored. If the path already has a descriptor use the overwrite flag to overwrite it")
	cmd.Flags().StringVar(&o.RegistryURL, "registry", "", "Repository context url for module to upload. The repository url will be automatically added to the repository contexts in the module")
	cmd.Flags().StringVarP(&o.Credentials, "credentials", "c", "", "Basic authentication credentials for the given registry in the format user:password")
	cmd.Flags().StringVarP(&o.TemplateOutput, "output", "o", "template.yaml", "File to output the module template if the module is uploaded to a registry")
	cmd.Flags().StringVar(&o.Channel, "channel", "stable", "Channel to use for the module template.")
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

	ref, err := oci.ParseRef(args[0])
	if err != nil {
		return err
	}

	if err := module.ValidateName(ref.ShortName()); err != nil {
		return err
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

	// Create base resource defs with module root and its sub-layers
	defs, err := module.Inspect(args[2], l)
	if err != nil {
		return err
	}

	// add extra resources
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

	/* -- PUSH & TEMPLATE -- */

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

		c.NewStep("Generating module template")
		t, err := module.Template(archive, c.opts.Channel, args[2], fs)
		if err != nil {
			c.CurrentStep.Failure()
			return err
		}

		if err := vfs.WriteFile(fs, c.opts.TemplateOutput, t, os.ModePerm); err != nil {
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
