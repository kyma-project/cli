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
		Use:   "module --name MODULE_NAME --version MODULE_VERSION --registry MODULE_REGISTRY [flags]",
		Short: "Signs all module resources from an unsigned component descriptor that's hosted in a remote OCI registry",
		Long: `Use this command to sign a Kyma module.

### Detailed description

This command signs all module resources recursively based on an unsigned component descriptor hosted in an OCI registry with the provided private key. Then, the output (component-descriptor.yaml) is saved in the descriptor path (default: ./mod) as a signed component descriptor. If signed-registry are provided, the command pushes the signed component descriptor.
`,
		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}
	cmd.Flags().StringVar(
		&o.Name, "name", "", "Name of the module.",
	)
	cmd.Flags().StringVar(
		&o.Version, "version", "", "Version of the module.",
	)
	cmd.Flags().StringVar(
		&o.PrivateKeyPath, "key", "", "Specifies the path where a private key is used for signing.",
	)
	cmd.Flags().StringVar(
		&o.RegistryURL, "registry", "", "Context URL of the repository for the module. "+
			"The repository's URL is automatically added to the repository's contexts in the module.",
	)
	cmd.Flags().StringVar(
		&o.NameMappingMode, "name-mapping", "urlPath",
		"Overrides the OCM Component Name Mapping, Use: \"urlPath\" or \"sha256-digest\".",
	)
	cmd.Flags().StringVarP(
		&o.Credentials, "credentials", "c", "",
		"Basic authentication credentials for the given registry in the user:password format",
	)
	cmd.Flags().StringVarP(
		&o.Token, "token", "t", "",
		"Authentication token for the given registry (alternative to basic authentication).",
	)
	cmd.Flags().BoolVar(&o.Insecure, "insecure", false, "Uses an insecure connection to access the registry.")

	return cmd
}

func (c *command) Run(_ []string) error {
	if !c.opts.NonInteractive {
		cli.AlphaWarn()
	}

	signCfg := &module.ComponentSignConfig{
		Name:    c.opts.Name,
		Version: c.opts.Version,
		KeyPath: c.opts.PrivateKeyPath,
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

	if err = module.Sign(signCfg, remote); err != nil {
		c.CurrentStep.Failure()
		return err
	}
	c.CurrentStep.Success()

	return nil
}
