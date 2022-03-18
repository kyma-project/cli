package get

import (
	"os"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/cmd/kyma/get/schema"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

var (
	refMap = map[string]func() ([]byte, error){
		"serverless": workspace.ReflectSchema,
	}
)

//NewCmd creates a new function command
func NewCmd(o *cli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Gets Kyma-related resources.",
		Long:  "Use this command to get Kyma-related resources.",
	}

	cmd.AddCommand(schema.NewCmd(schema.NewOptions(o, os.Stdout, refMap)))
	return cmd
}
