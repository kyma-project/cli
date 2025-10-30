# kyma alpha diagnose logs

Aggregate error logs from pods belonging to enabled Kyma Modules.

## Synopsis

Some better long description

```bash
kyma alpha diagnose logs [flags]
```

## Flags

```text
  -f, --format string           Output format (possible values: json, yaml)
      --lines int64             Max lines per container (default "200")
      --module stringSlice      Restrict to specific module(s). Can be used multiple times (default "[]")
  -o, --output string           Path to the diagnostic output file. If not provided the output is printed to stdout
      --since duration          Log time range (e.g., 10m, 1h, 30s) (default "0s")
      --timeout duration        Timeout for log collection operations (default "30s")
      --verbose                 Display verbose output, including error details during diagnostics collection
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha diagnose](kyma_alpha_diagnose.md) - Diagnose cluster health and configuration
