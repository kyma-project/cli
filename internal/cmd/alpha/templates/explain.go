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

func buildExplainCommand(out io.Writer, explainOptions *ExplainOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "explain",
		Short: explainOptions.Description,
		Long:  explainOptions.DescriptionLong,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintln(out, explainOptions.Output)
		},
	}
}
