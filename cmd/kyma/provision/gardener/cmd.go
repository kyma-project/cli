package gardener

import (
	"github.com/spf13/cobra"
)

// NewCmd creates a new gardener command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gardener",
		Short: "Provisions a cluster using Gardener on GCP, Azure, or AWS.",
	}
	return cmd
}
