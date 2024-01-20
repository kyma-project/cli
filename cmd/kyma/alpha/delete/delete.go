package delete

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/delete/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"disable"},
		Short:   "Disables a resource in the Kyma cluster.",
		Long: `Use this command to disable a resource in the Kyma cluster.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
