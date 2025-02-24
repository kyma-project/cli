package cmd

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha"
	"github.com/spf13/cobra"
)

func NewKymaCMD() (*cobra.Command, clierror.Error) {
	cmd := &cobra.Command{
		Use:   "kyma",
		Short: "Simple set of commands to manage a Kyma cluster",
		// Affects children as well
		// by default Cobra adds `Error:` to the front of the error message, we want to suppress it
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, _ []string) {
			if err := cmd.Help(); err != nil {
				_ = err
			}
		},
	}
	cmd.PersistentFlags().BoolP("help", "h", false, "Help for the command")

	alpha, err := alpha.NewAlphaCMD()
	if err != nil {
		return nil, err
	}
	cmd.AddCommand(alpha)

	return cmd, nil
}
