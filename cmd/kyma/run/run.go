package run

import (
	"github.com/kyma-project/cli/cmd/kyma/run/function"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new run command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run local resources.",
		Long:  `Use this command to run resources from the configuration.`,
	}

	cmd.AddCommand(function.NewCmd(function.NewOptions(o)))
	return cmd
}
