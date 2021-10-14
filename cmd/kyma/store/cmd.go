package store

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new store command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store",
		Short: "Stores certificates or host files in the local system.",
	}
	return cmd
}
