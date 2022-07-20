package verify

import (
	"github.com/kyma-project/cli/cmd/kyma/alpha/verify/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

//NewCmd verifys a new kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verifies kyma resources.",
		Long: `Use this command to verify Kyma resources.
`,
	}

	cmd.AddCommand(module.NewCmd(module.NewOptions(o)))

	return cmd
}
