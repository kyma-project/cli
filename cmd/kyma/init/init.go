package init

import (
	"github.com/kyma-project/cli/cmd/kyma/init/function"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new function command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Creates local resources for your project.",
		Long:  "Use this command to create the local files and folders for your project.",
	}

	cmd.AddCommand(function.NewCmd(function.NewOptions(o)))
	return cmd
}
