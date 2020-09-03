package apply

import (
	"github.com/kyma-project/cli/cmd/kyma/apply/function"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new function command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Applies local resources to the Kyma cluster.",
		Long:  "Use this command to apply the local files and folders from your project to the Kyma cluster.",
	}

	cmd.AddCommand(function.NewCmd(function.NewOptions(o)))
	return cmd
}
