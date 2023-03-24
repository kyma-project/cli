package module

import (
	"github.com/kyma-project/cli/internal/cli"
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
		Short: "Unpacks an OCI container image module bundled as an  from the contents of the given path",
		Long: `Use this command to unpack a Kyma module.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. Modules are built and distributed as OCI container images. 
With this command, you can unpack the contents of an image so that they can be deployed into a cluster or inspected by developers.
`,

		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"mod"},
	}

	return cmd
}

func (c *command) Run(_ []string) error {
	return nil
}
