package templates

import (
	"fmt"
	"io"
	"os"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/spf13/cobra"
)

type ExplainOptions struct {
	types.ExplainCommand
}

func BuildExplainCommand(explainOptions *ExplainOptions) *cobra.Command {
	return buildExplainCommand(os.Stdout, explainOptions)
}

func buildExplainCommand(out io.Writer, options *ExplainOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "explain [flags]",
		Short: options.Description,
		Long:  options.DescriptionLong,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintln(out, options.Output)
		},
	}
}
