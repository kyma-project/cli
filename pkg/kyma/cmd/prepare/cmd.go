package prepare

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new prepare command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepares a cluster for installation",
	}
	return cmd
}
