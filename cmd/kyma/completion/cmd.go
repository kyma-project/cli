package completion

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

//NewCmd creates a new completion command
func NewCmd() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion bash|zsh",
		Short: "Generates bash or zsh completion scripts.",
		Long: `Use this command to display the shell completion code used for interactive command completion. 
To configure your shell to load completions, add ` + "`. <(kyma completion bash)`" + ` to your bash profile or ` + "`. <(kyma completion zsh)`" + ` to your zsh profile.
`,
		RunE:    completion,
		Aliases: []string{},
	}
	return completionCmd
}

func completion(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		fmt.Println("Usage: kyma completion bash|zsh")
		fmt.Println("See 'kyma completion -h' for help")
		return nil
	}

	switch shell := args[0]; shell {
	case "bash":
		err := cmd.GenBashCompletion(os.Stdout)
		return errors.Wrap(err, "Error generating bash completion")
	case "zsh":
		err := genZshCompletion(cmd, os.Stdout)
		return errors.Wrap(err, "Error generating zsh completion")
	default:
		fmt.Printf("Sorry, completion is not supported for %q", shell)
	}

	return nil
}

func genZshCompletion(cmd *cobra.Command, out io.Writer) error {
	err := cmd.GenZshCompletion(out)
	if err != nil {
		return err
	}

	zshCompdef := "\ncompdef _kyma kyma\n"

	_, err = io.WriteString(out, zshCompdef)
	return err
}
