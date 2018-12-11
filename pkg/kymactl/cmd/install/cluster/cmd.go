package cluster

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new cluster command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Prepares a cluster for installation",
	}
	return cmd
}
