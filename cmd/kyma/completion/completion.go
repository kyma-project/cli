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
		Long: `Use this command to display the shell completion code for bash. The shell code must be evaluated to provide
interactive completion of commands. To do this, source it from the bash profile.
To load completion, run:

    . <(kyma completion)

To configure your bash shell to load completions for each session, add ` + "`. <(kyma completion)`" + ` to your bash profile.

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
