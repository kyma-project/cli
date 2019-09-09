package completion

import (
	"os"

	"github.com/spf13/cobra"
)

//NewCompletionCmd creates a new completion command
func NewCmd() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts.",
		Long: `Use this command to display the shell completion code used for interactive command completion. 
To configure your bash shell to load completions, add ` + "`. <(kyma completion)`" + ` to your bash profile.

### Usage 
`,
		RunE:    completion,
		Aliases: []string{},
	}
	return completionCmd
}

func completion(cmd *cobra.Command, args []string) error {
	err := cmd.GenBashCompletion(os.Stdout)
	return err
}
