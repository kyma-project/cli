package create

import (
	"os"

	"github.com/kyma-project/cli/cmd/kyma/create/schema"
	"github.com/kyma-project/cli/cmd/kyma/create/system"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

var (
	refMap = map[string]func() ([]byte, error){}
)

//NewCmd creates a new kyma CLI command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates resources on the Kyma cluster.",
		Long: `Use this command to create resources on the Kyma cluster.
`,
	}

	cmd.AddCommand(system.NewCmd(system.NewOptions(o)))
	cmd.AddCommand(schema.NewCmd(schema.NewOptions(o, os.Stdout, refMap)))

	return cmd
}
