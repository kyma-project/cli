package version

import (
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
)

var (
	version = "local"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Displays the version of Kyma CLI",
		Long:  "Use this command to print the version of Kyma CLI.",
		Run: func(cmd *cobra.Command, args []string) {
			out.Msgfln("Kyma-CLI Version: %s", version)
		},
	}
}
