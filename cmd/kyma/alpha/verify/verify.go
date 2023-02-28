package verify

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/verify/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCmd verifys a new kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verifies all module resources from a signed module component descriptor that's hosted in a remote OCI registry",
		Long:  "Use this command to verify all module resources from a signed module descriptor that's hosted in a remote OCI registry",
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
