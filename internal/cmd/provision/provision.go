package provision

import "github.com/spf13/cobra"

func NewProvisionCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use: "provision",
	}

	return cmd
}
