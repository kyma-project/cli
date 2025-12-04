# kyma alpha diagnose

Runs diagnostic commands to troubleshoot your Kyma cluster.

## Synopsis

Use diagnostic commands to collect cluster information, analyze logs, and assess the health of your Kyma installation. Choose from available subcommands to target specific diagnostic areas.

```bash
kyma alpha diagnose <command> [flags]
```

## Available Commands

```text
  cluster - Diagnoses cluster health and configuration
  istio   - Checks Istio configuration
  logs    - Aggregates error logs from Pods belonging to the added Kyma modules (requires JSON log format)
```

## Flags

```text
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha](kyma_alpha.md)                                   - Groups command prototypes for which the API may still change
* [kyma alpha diagnose cluster](kyma_alpha_diagnose_cluster.md) - Diagnoses cluster health and configuration
* [kyma alpha diagnose istio](kyma_alpha_diagnose_istio.md)     - Checks Istio configuration
* [kyma alpha diagnose logs](kyma_alpha_diagnose_logs.md)       - Aggregates error logs from Pods belonging to the added Kyma modules (requires JSON log format)
