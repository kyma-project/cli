package enable

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/enable/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enables a resource on the Kyma cluster.",
		Long: `Use this command to enable a resource on the Kyma cluster.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
