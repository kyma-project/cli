package cluster

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Prepares a cluster for installation",
	}
	return cmd
}
