package install

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "kyma installation",
	}
	return cmd
}
