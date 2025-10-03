# kyma module list

Lists the installed modules.

## Synopsis

Use this command to list the installed Kyma modules.

```bash
kyma module list [flags]
```

## Flags

```text
  -o, --output string           Output format (Possible values: table, json, yaml)
      --show-errors             Indicates whether to show errors outputted by misconfigured modules
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
  -q, --quiet                   Suppress non-essential output
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
