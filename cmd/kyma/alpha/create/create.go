package create

import (
	"github.com/spf13/cobra"

	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
	"github.com/kyma-project/cli/internal/cli"
)

// NewCmd creates a new Kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates resources on the Kyma cluster.",
		Long: `Use this command to create resources on the Kyma cluster.
`,
	}

	c := module.NewCmd(module.NewOptions(o)).Command
	cmd.AddCommand(&c)

	return cmd
}
