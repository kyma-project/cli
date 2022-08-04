package alpha

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/create"
	initcmd "github.com/kyma-project/cli/cmd/kyma/alpha/init"
	"github.com/kyma-project/cli/cmd/kyma/alpha/sign"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new Kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha",
		Short: "Experimental commands",
		Long: `Alpha commands are experimental, unreleased features that should only be used by the Kyma team. Use at your own risk.
`,
	}

	cmd.AddCommand(create.NewCmd(o))
	cmd.AddCommand(sign.NewCmd(o))
	cmd.AddCommand(initcmd.NewCmd(o))

	return cmd
}
