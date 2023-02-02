package disable

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/disable/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disables a resource on the Kyma cluster.",
		Long: `Use this command to disable a resource on the Kyma cluster.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
