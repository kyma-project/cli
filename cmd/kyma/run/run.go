package run

import (
	"github.com/kyma-project/cli/cmd/kyma/run/function"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new run command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
	}

	cmd.AddCommand(function.NewCmd(function.NewOptions(o)))
	return cmd
}
