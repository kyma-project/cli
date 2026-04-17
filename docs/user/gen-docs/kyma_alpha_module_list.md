# kyma alpha module list

Lists installed modules.

## Synopsis

Use this command to list the installed Kyma modules.

NOTE: functionality under construction
  - community modules not yet supported

```bash
kyma alpha module list [flags]
```

## Flags

```text
  -o, --output string           Output format (Possible values: table, json, yaml)
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha module](kyma_alpha_module.md) - Manages Kyma modules
