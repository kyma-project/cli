package cmd

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha"
	"github.com/kyma-project/cli.v3/internal/cmd/version"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewKymaCMD() (*cobra.Command, clierror.Error) {
	cmd := &cobra.Command{
		Use:   "kyma <command> [flags]",
		Short: "A simple set of commands to manage a Kyma cluster",
		Long:  "Use this command to manage Kyma modules and resources on a cluster.",
		// Affects children as well
		// by default Cobra adds `Error:` to the front of the error message, we want to suppress it
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmdcommon.AddExtensionsFlags(cmd)
	cmd.PersistentFlags().BoolP("help", "h", false, "Help for the command")

	alpha, err := alpha.NewAlphaCMD()
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(alpha)
	cmd.AddCommand(version.NewCmd())

	return cmd, nil
}
