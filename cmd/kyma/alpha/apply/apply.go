package apply

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/apply/function"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new function command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Applies local resources to the Kyma cluster.",
		Long:  "Use this command to apply the resource configuration to the Kyma cluster.",
	}

	cmd.AddCommand(function.NewCmd(function.NewOptions(o)))
	return cmd
}
