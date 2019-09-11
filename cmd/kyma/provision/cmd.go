package provision

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new provision command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provisions a cluster for Kyma installation.",
	}
	cmd.Flags().Bool("help", false, "Displays help for the command.")
	return cmd
}
