package function

import (
	"github.com/kyma-project/cli/cmd/kyma/function/apply"
	_init "github.com/kyma-project/cli/cmd/kyma/function/init"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new function command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "function",
		Short: "Manage functions.",
	}
	cmd.AddCommand(apply.NewCmd(apply.NewOptions(o)))
	cmd.AddCommand(_init.NewCmd(_init.NewOptions(o)))
	return cmd
}
