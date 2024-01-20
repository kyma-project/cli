package create

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
	"github.com/kyma-project/cli/cmd/kyma/alpha/create/scaffold"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCmd creates a new Kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates resources on the Kyma cluster.",
		Long: `Use this command to create resources on the Kyma cluster.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))
	cmd.AddCommand(scaffold.NewCmd(scaffold.NewOptions(o)))

	return cmd
}
