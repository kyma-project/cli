package alpha

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new alpha command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha",
		Short: "Executes the commands in alpha testing stage.",
	}
	return cmd
}
