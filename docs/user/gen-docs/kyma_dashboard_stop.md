# kyma dashboard stop

Terminates the locally running Kyma dashboard.

## Synopsis

Use this command to terminate the locally running Kyma dashboard in a Docker container.

```bash
kyma dashboard stop [flags]
```

## Flags

```text
      --container-name string   Specifies the name of the local container to stop. (default "kyma-dashboard")
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma dashboard](kyma_dashboard.md) - Manages Kyma dashboard locally.
