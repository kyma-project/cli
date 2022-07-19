package module

import (
	"github.com/kyma-project/cli/internal/cli"
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
		Use:   "module <MODULE_IMAGE>",
		Short: "Verifies the signature of a Kyma module bundled as an OCI container image.",
		Long: `Use this command to verify a Kyma module.

### Detailed description

Kyma modules can be cryptographically signed to make sure they are correct and distributed by a trusted authority. This command verifies the authenticity of a given module.
`,

		RunE: func(_ *cobra.Command, args []string) error { return c.Run(args) },
	}

	return cmd
}

func (c *command) Run(args []string) error {
	return nil
}
