# kyma completion powershell

Generate the autocompletion script for powershell.

## Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	kyma completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```bash
kyma completion powershell [flags]
```

## Flags

```text
      --no-descriptions         disable completion descriptions
  -h, --help                    Help for the command
      --show-extensions-error   Prints a possible error when fetching extensions fails
```

## See also

* [kyma completion](kyma_completion.md) - Generate the autocompletion script for the specified shell
