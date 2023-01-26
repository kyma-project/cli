---
title: kyma completion
---

Generates bash or zsh completion scripts.

## Synopsis

Use this command to display the shell completion code used for interactive command completion. 
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


```bash
kyma completion bash|zsh [flags]
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner).
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](kyma.md)	 - Controls a Kyma cluster.

