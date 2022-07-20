package module

import (
	"fmt"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/module"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"path/filepath"
	"sigs.k8s.io/yaml"
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
		Short: "Sign all module resources from unsigned component descriptor which hosted in a remote OCI registry",
		Long: `Use this command to sign a Kyma module.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. Modules are built and distributed as OCI container images. 

This command signing all module resources from unsigned component descriptor in a remote OCI registry.
`,
		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}
	cmd.Args = cobra.ExactArgs(2)
	cmd.Flags().StringVar(&o.PrivateKeyPath, "private-key", "", "Specifies the path where the private key used for signing")
	cmd.Flags().StringVar(&o.SignatureName, "signature-name", "", "name of the signature for signing")
	cmd.Flags().StringVar(&o.RegistryURL, "registry", "", "Repository context url where unsigned component descriptor located.")
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
		RegistryURL:    c.opts.RegistryURL,
		PrivateKeyPath: c.opts.PrivateKeyPath,
	}
	cfg := &module.ComponentConfig{
		Name:                 args[0],
		Version:              args[1],
		ComponentArchivePath: "./mod",
		Overwrite:            true,
		RegistryURL:          c.opts.RegistryURL,
	}

	remote := &module.Remote{
		Registry:    c.opts.RegistryURL,
		Credentials: c.opts.Credentials,
		Token:       c.opts.Token,
		Insecure:    c.opts.Insecure,
	}
	fs := osfs.New()

	c.NewStep(fmt.Sprintf("Creating module archive at %q", "./mod"))
	archive, err := module.Build(fs, cfg)
	if err != nil {
		c.CurrentStep.Failure()
		return err
	}
	c.NewStep("Signing resources...")
	digestedCds, err := module.Sign(archive, signCfg, c.opts.PrivateKeyPath, c.opts.SignatureName, remote, log)
	if err != nil {
		c.CurrentStep.Failure()
		return err
	}
	compDescFilePath := filepath.Join(cfg.ComponentArchivePath, ctf.ComponentDescriptorFileName)
	for _, digestedCd := range digestedCds {
		c.CurrentStep.Successf("Signed component descriptor")
		data, err := yaml.Marshal(digestedCd)
		if err != nil {
			return fmt.Errorf("unable to encode component descriptor: %w", err)
		}
		if err := vfs.WriteFile(fs, compDescFilePath, data, 0664); err != nil {
			return fmt.Errorf("unable to write modified comonent descriptor: %w", err)
		}
	}
	return nil
}
