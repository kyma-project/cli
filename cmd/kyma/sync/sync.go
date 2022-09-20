package sync

import (
	"github.com/kyma-project/cli/cmd/kyma/sync/function"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCmd creates a new function command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronizes the local resources for your Function.",
		Long:  "Use this command to download the given resource's code and dependencies from the cluster to create or update these resources in your local workspace. Currently, you can only use it for Functions.",
	}

	cmd.AddCommand(function.NewCmd(function.NewOptions(o)))
	return cmd
}
