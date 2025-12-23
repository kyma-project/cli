# kyma dashboard start

Runs Kyma dashboard locally.

## Synopsis

Use this command to run Kyma dashboard locally in a Docker container and open it directly in a web browser.

```bash
kyma dashboard start [flags]
```

## Flags

```text
      --container-name string   Specify the name of the local container. (default "kyma-dashboard")
  -o, --open                    Specify if the browser should open after executing the command.
  -p, --port string             Specify the port on which the local dashboard will be exposed. (default "3001")
  -v, --verbose                 Enable verbose output with detailed logs.
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma dashboard](kyma_dashboard.md) - Manages Kyma dashboard locally.
