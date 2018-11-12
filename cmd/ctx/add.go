package ctx

import (
	"github.com/spf13/cobra"
)

func newCmdCtxAdd() *cobra.Command {
	ctxAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new context or replace an existing one",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO add context to the config
		},
	}
	return ctxAddCmd
}
