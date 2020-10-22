package sync

import (
	"github.com/kyma-project/cli/cmd/kyma/sync/function"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new function command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronizes the local resources for your Function.",
		Long:  "Use this command to synchronise the function files from the cluster.",
	}

	cmd.AddCommand(function.NewCmd(function.NewOptions(o)))
	return cmd
}
