package module

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module"
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
		Use:   "module <MODULE_IMAGE>",
		Short: "Verifies the signature of a Kyma module bundled as an OCI container image.",
		Long: `Use this command to verify a Kyma module.

### Detailed description

Kyma modules can be cryptographically signed to make sure they are correct and distributed by a trusted authority. This command verifies the authenticity of a given module.
`,

		RunE: func(_ *cobra.Command, args []string) error { return c.Run(args) },
	}

	cmd.Flags().StringVar(
		&o.Name, "name", "", "Name of the module",
	)
	cmd.Flags().StringVar(
		&o.Version, "version", "", "Version of the module",
	)
	cmd.Flags().StringVar(
		&o.PublicKeyPath, "key", "", "Specifies the path where the private key used for signing",
	)
	cmd.Flags().StringVar(&o.SignatureName, "signature-name", "", "name of the signature for signing")
	cmd.Flags().StringVar(
		&o.RegistryURL, "registry", "", "Repository context url where unsigned component descriptor located",
	)
	cmd.Flags().StringVar(
		&o.NameMappingMode, "nameMapping", "urlPath",
		"Overrides the OCM Component Name Mapping, one of: \"urlPath\" or \"sha256-digest\"",
	)
	cmd.Flags().StringVarP(
		&o.Credentials, "credentials", "c", "",
		"Basic authentication credentials for the given registry in the format user:password",
	)
	cmd.Flags().StringVarP(
		&o.Token, "token", "t", "",
		"Authentication token for the given registry (alternative to basic authentication).",
	)
	cmd.Flags().BoolVar(&o.Insecure, "insecure", false, "Use an insecure connection to access the registry.")

	return cmd
}

func (c *command) Run(_ []string) error {
	if !c.opts.NonInteractive {
		cli.AlphaWarn()
	}

	signCfg := &module.ComponentSignConfig{
		Name:          c.opts.Name,
		Version:       c.opts.Version,
		KeyPath:       c.opts.PublicKeyPath,
		SignatureName: c.opts.SignatureName,
	}

	c.NewStep("Fetching and signing component descriptor...")
	nameMappingMode, err := module.ParseNameMapping(c.opts.NameMappingMode)
	if err != nil {
		c.CurrentStep.Failure()
		return err
	}

	remote := &module.Remote{
		Registry:    c.opts.RegistryURL,
		NameMapping: nameMappingMode,
		Credentials: c.opts.Credentials,
		Token:       c.opts.Token,
		Insecure:    c.opts.Insecure,
	}

	if err := module.Verify(signCfg, remote); err != nil {
		c.CurrentStep.Failure()
		return err
	}
	c.CurrentStep.Success()

	return nil
}
