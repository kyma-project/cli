package templates

import (
	"fmt"

	"github.com/spf13/cobra"
)

type ExplainOptions struct {
	Short  string
	Long   string
	Output string
}

func BuildExplainCommand(explainOptions *ExplainOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "explain",
		Short: explainOptions.Short,
		Long:  explainOptions.Long,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(explainOptions.Output)
		},
	}
}
