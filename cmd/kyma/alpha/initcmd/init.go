package initcmd

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/initcmd/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new Kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initializes resources locally.",
		Long: `Use this command to initialize resources locally.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
