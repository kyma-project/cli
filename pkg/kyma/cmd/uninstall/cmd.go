package uninstall

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new uninstall command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "kyma uninstallation",
	}
	return cmd
}
