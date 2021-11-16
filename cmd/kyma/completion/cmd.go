package completion

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewCmd creates a new completion command
func NewCmd() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion bash|zsh",
		Short: "Generates bash or zsh completion scripts.",
		Long: `Use this command to display the shell completion code used for interactive command completion. 
To configure your shell to load completions, use one of the following commands:
for the bash: ". <(kyma completion bash)",
for the zsh: ". <(kyma completion zsh)",
for the fish: "kyma completion fish | source",
for the powershell: "kyma completion powershell | Out-String | Invoke-Expression".
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
		err := cmd.Root().GenBashCompletion(os.Stdout)
		return errors.Wrap(err, "Error generating bash completion")
	case "zsh":
		err := genZshCompletion(cmd, os.Stdout)
		return errors.Wrap(err, "Error generating zsh completion")
	case "fish":
		err := cmd.Root().GenFishCompletion(os.Stdout, false)
		return errors.Wrap(err, "Error generating fish completion")
	case "powershell":
		err := cmd.Root().GenPowerShellCompletion(os.Stdout)
		return errors.Wrap(err, "Error generating powershell completion")
	default:
		fmt.Printf("Sorry, completion is not supported for %q", shell)
	}

	return nil
}

func genZshCompletion(cmd *cobra.Command, out io.Writer) error {
	err := cmd.Root().GenZshCompletion(out)
	if err != nil {
		return err
	}

	zshCompdef := "\ncompdef _kyma kyma\n"

	_, err = io.WriteString(out, zshCompdef)
	return err
}
