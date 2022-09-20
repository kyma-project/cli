package unpack

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCmd creates a new Kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpack",
		Short: "Unpack Kyma resources to the local filesystem.",
		Long: `Use this command to unpack Kyma resources  to the local filesystem.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
