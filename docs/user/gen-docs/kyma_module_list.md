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
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
