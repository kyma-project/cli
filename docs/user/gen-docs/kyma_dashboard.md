# kyma dashboard

Manages Kyma dashboard locally.

## Synopsis

Use this command to manage Kyma dashboard locally in a Docker container.

```bash
kyma dashboard <command> [flags]
```

## Available Commands

```text
  start - Runs Kyma dashboard locally.
  stop  - Terminates the locally running Kyma dashboard.
```

## Flags

```text
      --container-name string   Specifies the name of the local container. (default "kyma-dashboard")
  -p, --port string             Specifies the port on which the local dashboard will be exposed. (default "8000")
  -v, --verbose                 Enables verbose output with detailed logs.
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma](kyma.md)                                 - A simple set of commands to manage a Kyma cluster
* [kyma dashboard start](kyma_dashboard_start.md) - Runs Kyma dashboard locally.
* [kyma dashboard stop](kyma_dashboard_stop.md)   - Terminates the locally running Kyma dashboard.
