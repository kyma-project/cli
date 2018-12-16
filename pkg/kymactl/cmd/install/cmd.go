package install

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new install command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "kyma installation",
	}
	return cmd
}
