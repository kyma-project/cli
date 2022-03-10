package create

import (
	"github.com/kyma-project/cli/cmd/kyma/create/system"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates resources on the Kyma cluster.",
		Long: `Use this command to create resources on the Kyma cluster.
`,
	}

	cmd.AddCommand(system.NewCmd(system.NewOptions(o)))

	return cmd
}
