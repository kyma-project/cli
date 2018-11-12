package ctx

import (
	"fmt"
	"os"

	"github.com/kyma-incubator/kymactl/config"
	"github.com/spf13/cobra"
)

func newCmdCtxRm() *cobra.Command {
	ctxRmCmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove a context",
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				// TODO improve this, do we really want to error? print usage!
				fmt.Println("Context does not exist")
				os.Exit(1)
			case 1:
				deleteCtx(args[0])
			}

		},
		Aliases: []string{"remove"},
	}
	return ctxRmCmd
}

func deleteCtx(ctxName string) {
	cfg, err := config.Context()
	if err != nil {
		fmt.Printf("Error getting context configuration: %s\n", err)
		os.Exit(1)
	}
	delete(cfg.Contexts, ctxName)

	if err = config.SaveContext(cfg); err != nil {
		fmt.Printf("Error saving context configuration: %s\n", err)
		os.Exit(1)
	}
}
