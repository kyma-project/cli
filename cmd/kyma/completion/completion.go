package completion

import (
	"os"

	"github.com/spf13/cobra"
)

//NewCompletionCmd creates a new completion command
func NewCmd() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: `Output shell completion code for bash. The shell code must be evaluated to provide
interactive completion of commands. This can be done by sourcing it from the .bash _profile.
To load completion, run:

. <(kyma completion)

To configure your bash shell to load completions for each session, add to your bashrc:

# ~/.bashrc or ~/.profile
. <(kyma completion)
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
