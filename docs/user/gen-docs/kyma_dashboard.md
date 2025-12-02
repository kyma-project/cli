# kyma dashboard

Manage Kyma dashboard locally.

## Synopsis

Use this command to manage running the Kyma dashboard locally in a docker container.

```bash
kyma dashboard <command> [flags]
```

## Available Commands

```text
  start - Run Kyma dashboard locally
  stop  - Run Kyma dashboard locally
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

* [kyma](kyma.md)                                 - A simple set of commands to manage a Kyma cluster
* [kyma dashboard start](kyma_dashboard_start.md) - Run Kyma dashboard locally
* [kyma dashboard stop](kyma_dashboard_stop.md)   - Run Kyma dashboard locally
