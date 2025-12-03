# kyma alpha diagnose istio

Checks Istio configuration.

## Synopsis

Use this command to quickly assess potential Istio configuration issues in your cluster for troubleshooting and support purposes.

```bash
kyma alpha diagnose istio [flags]
```

## Examples

```bash
  # Analyze Istio configuration across all namespaces
  kyma alpha diagnose istio
  # or
  kyma alpha diagnose istio --all-namespaces

  # Analyze Istio configuration in a specific namespace
  kyma alpha diagnose istio --namespace my-namespace

  # Print only warnings and errors
  kyma alpha diagnose istio --level warning

  # Output as JSON to a file
  kyma alpha diagnose istio --format json --output istio-diagnostics.json
```

## Flags

```text
  -A, --all-namespaces          Analyzes all namespaces
  -f, --format string           Output format (possible values: json, yaml)
      --level string            Output message level (possible values: info, warning, error) (default "warning")
  -n, --namespace string        The namespace that the workload instances belongs to
  -o, --output string           Path to the diagnostic output file. If not provided the output is printed to stdout
      --timeout duration        Timeout for diagnosis (default "30s")
      --verbose                 Displays verbose output, including error details during diagnostics collection
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha diagnose](kyma_alpha_diagnose.md) - Diagnose cluster health and configuration
