# kyma alpha diagnose logs

Aggregates error logs from Pods belonging to the added Kyma modules.

## Synopsis

Collects and aggregates recent error-level container logs for Kyma modules to help with rapid troubleshooting.

EXAMPLES:
  # Collect last 200 lines (default) from all enabled modules
  kyma alpha diagnose logs --lines 200

  # Collect error logs from the last 15 minutes for all enabled modules
  kyma alpha diagnose logs --since 15m

  # Restrict to specific modules (repeat --module) and increase line count
  kyma alpha diagnose logs --module serverless --module api-gateway --lines 500

  # Time-based collection for one module, output as JSON to a file
  kyma alpha diagnose logs --module serverless --since 30m --format json --output serverless-errors.json

  # Collect with verbose output and shorter timeout (useful in CI)
  kyma alpha diagnose logs --since 10m --timeout 10s --verbose

  # Use lines as a deterministic cap when time window would be too large
  kyma alpha diagnose logs --lines 1000

NOTE: --since takes precedence over --lines when both are provided; use only one for clarity.

```bash
kyma alpha diagnose logs [flags]
```

## Flags

```text
  -f, --format string           Output format (possible values: json, yaml)
      --lines int64             Max lines per container (default "200")
      --module stringSlice      Restrict to specific module(s). Can be used multiple times. When no value is present then logs from all Kyma CR modules are gathered (default "[]")
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
