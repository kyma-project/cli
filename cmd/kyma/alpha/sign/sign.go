package sign

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/sign/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd creates a new Kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign all module resources from unsigned component descriptor which hosted in a remote OCI registry",
		"Use this command to sign all module resources from an unsigned component descriptor that's hosted in a remote OCI registry",
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
