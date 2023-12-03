package module

import (
	"github.com/spf13/cobra"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module"
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
		Short: "Verifies the signature of a Kyma module bundled as an OCI container image.",
		Long: `Use this command to verify a Kyma module.

### Detailed description

Kyma modules can be cryptographically signed to ensure they are correct and distributed by a trusted authority. This command verifies the authenticity of a given module.
`,

		RunE: func(_ *cobra.Command, args []string) error { return c.Run(args) },
	}

	cmd.Flags().StringVar(
		&o.Name, "name", "", "Name of the module.",
	)
	cmd.Flags().StringVar(
		&o.Version, "version", "", "Version of the module.",
	)
	cmd.Flags().StringVar(
		&o.PublicKeyPath, "key", "", "Specifies the path where a public key is used for signing.",
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
		KeyPath: c.opts.PublicKeyPath,
	}

	c.NewStep("Fetching and verifying component descriptor...")
	nameMappingMode, err := module.ParseNameMapping(c.opts.NameMappingMode)
	if err != nil {
		return err
	}

	remote := &module.Remote{
		Registry:      c.opts.RegistryURL,
		NameMapping:   nameMappingMode,
		Credentials:   c.opts.Credentials,
		Token:         c.opts.Token,
		Insecure:      c.opts.Insecure,
		OciRepoAccess: &module.OciRepo{},
	}

	if err := module.Verify(signCfg, remote); err != nil {
		c.CurrentStep.Failuref("Invalid!")
		return err
	}
	c.CurrentStep.Successf("Valid!")

	return nil
}
