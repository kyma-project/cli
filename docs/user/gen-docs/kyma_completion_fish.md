# kyma completion fish

Generate the autocompletion script for fish.

## Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	kyma completion fish | source

To load completions for every new session, execute once:

	kyma completion fish > ~/.config/fish/completions/kyma.fish

You will need to start a new shell for this setup to take effect.


```bash
kyma completion fish [flags]
```

## Flags

```text
      --no-descriptions         disable completion descriptions
  -h, --help                    Help for the command
      --show-extensions-error   Prints a possible error when fetching extensions fails
```

## See also

* [kyma completion](kyma_completion.md) - Generate the autocompletion script for the specified shell
