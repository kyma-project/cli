package provision

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new provision command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "Provisions",
		Short: "Provisions a cluster for installation",
	}
	return cmd
}
