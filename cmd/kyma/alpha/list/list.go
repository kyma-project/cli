package list

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/list/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCmd creates a new Kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"get"},
		Short:   "Lists resources on the Kyma cluster.",
		Long: `Use this command to list resources on the Kyma cluster.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
