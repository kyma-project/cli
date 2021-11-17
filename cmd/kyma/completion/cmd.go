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
To configure your shell to load completions, use:

Bash:

  $ source <(kyma completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ kyma completion bash > /etc/bash_completion.d/kyma
  # macOS:
  $ kyma completion bash > /usr/local/etc/bash_completion.d/kyma

Zsh:

  $ source <(kyma completion zsh)

  # To load completions for each session, execute once:
  $ kyma completion zsh > "${fpath[1]}/_kyma"

  # You will need to start a new shell for this setup to take effect.

  # If shell completion is not already enabled in your environment, you must enable it.
  # Execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

Fish:

  $ kyma completion fish | source

  # To load completions for each session, execute once:
  $ kyma completion fish > ~/.config/fish/completions/kyma.fish

Powershell:

  PS> kyma completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> kyma completion powershell > kyma.ps1
  # and source this file from your PowerShell profile.
`,
		RunE:    completion,
		Aliases: []string{},
	}
	return completionCmd
}

func completion(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		fmt.Println("Usage: kyma completion bash|zsh|fish|powershell")
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
		err := cmd.Root().GenFishCompletion(os.Stdout, true)
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
