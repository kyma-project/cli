package module

import (
	"fmt"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module"
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
		Use:   "module MODULE_NAME MODULE_VERSION [flags]",
		Short: "Sign all module resources from unsigned component descriptor which is hosted in a remote OCI registry",
		Long: `Use this command to sign a Kyma module.

### Detailed description

This command signs all module resources recursively based on an unsigned component descriptor hosted in an OCI registry with the provided private key. The output (component-descriptor.yaml), is then saved in the descriptor path (./mod as a default) as a signed component descriptor. If signed-registry are provided, it will push the signed component descriptor.
`,
		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}
	cmd.Args = cobra.ExactArgs(2)
	cmd.Flags().StringVar(&o.PrivateKeyPath, "private-key", "", "Specifies the path where the private key used for signing")
	cmd.Flags().StringVar(&o.ModPath, "mod-path", "./mod", "Specifies the path where the signed component descriptor will be stored")
	cmd.Flags().StringVar(&o.SignatureName, "signature-name", "", "name of the signature for signing")
	cmd.Flags().StringVar(&o.RegistryURL, "registry", "", "Repository context url where unsigned component descriptor located")
	cmd.Flags().StringVar(&o.SignedRegistryURL, "signed-registry", "", "Repository context url where signed component descriptor located")
	cmd.Flags().StringVarP(&o.Credentials, "credentials", "c", "", "Basic authentication credentials for the given registry in the format user:password")
	cmd.Flags().StringVarP(&o.Token, "token", "t", "", "Authentication token for the given registry (alternative to basic authentication).")
	cmd.Flags().BoolVar(&o.Insecure, "insecure", false, "Use an insecure connection to access the registry.")

	return cmd
}

func (c *command) Run(args []string) error {
	if !c.opts.NonInteractive {
		cli.AlphaWarn()
	}

	log := cli.NewLogger(c.opts.Verbose).Sugar()

	signCfg := &module.ComponentSignConfig{
		Name:           args[0],
		Version:        args[1],
		PrivateKeyPath: c.opts.PrivateKeyPath,
		SignatureName:  c.opts.SignatureName,
	}

	remote := &module.Remote{
		Registry:    c.opts.RegistryURL,
		Credentials: c.opts.Credentials,
		Token:       c.opts.Token,
		Insecure:    c.opts.Insecure,
	}

	c.NewStep("Fetching and signing component descriptor...")
	digestedCds, err := module.Sign(signCfg, remote, log)
	if err != nil {
		c.CurrentStep.Failure()
		return err
	}

	// TODO: at the moment only support one cd, consider extend this further
	if len(digestedCds) < 1 {
		c.CurrentStep.Failure()
	}
	c.CurrentStep.Success()

	c.NewStep("Generating signed component descriptor...")
	fs := osfs.New()
	firstDigestedCd := digestedCds[0]
	if err := module.WriteComponentDescriptor(fs, firstDigestedCd, c.opts.ModPath, ctf.ComponentDescriptorFileName); err != nil {
		c.CurrentStep.Failure()
		return err
	}

	c.CurrentStep.Successf("Signed component descriptor generated at %s", c.opts.ModPath)

	if c.opts.SignedRegistryURL != "" {
		cfg := &module.ComponentConfig{
			Name:                 signCfg.Name,
			Version:              signCfg.Version,
			ComponentArchivePath: c.opts.ModPath,
			Overwrite:            false,
			RegistryURL:          c.opts.SignedRegistryURL,
		}
		archive, err := module.Build(fs, cfg)
		if err != nil {
			c.CurrentStep.Failure()
			return err
		}
		c.CurrentStep.Success()

		c.NewStep(fmt.Sprintf("Pushing signed component descriptor to %q", c.opts.SignedRegistryURL))
		archive.ComponentDescriptor = firstDigestedCd
		// Assume the credential are same between registry which host unsigned and signed component descriptor
		remote.Registry = c.opts.SignedRegistryURL
		if err := module.Push(archive, remote, log); err != nil {
			c.CurrentStep.Failure()
			return err
		}
		c.CurrentStep.Success()
	}
	return nil
}
