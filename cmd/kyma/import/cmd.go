package imprt

import (
	"github.com/spf13/cobra"
)

// NewCmd creates a new import command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Imports certificates to local certificates storage or adds domains to the local host file.",
	}
	return cmd
}
