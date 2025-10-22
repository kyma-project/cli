# kyma alpha diagnose

Diagnose cluster health and configuration.

## Synopsis

Use this command to quickly assess the health, configuration, and potential issues in your cluster for troubleshooting and support purposes.

```bash
kyma alpha diagnose [flags]
```

## Flags

```text
  -f, --format string           Output format (possible values: json, yaml)
  -o, --output string           Path to the diagnostic output file. If not provided the output is printed to stdout
      --verbose                 Display verbose output, including error details during diagnostics collection
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha](kyma_alpha.md) - Groups command prototypes for which the API may still change
