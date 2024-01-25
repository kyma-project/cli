package add

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/add/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"enable"},
		Short:   "Adds a resource to the Kyma cluster.",
		Long: `Use this command to enable a resource in the Kyma cluster.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
