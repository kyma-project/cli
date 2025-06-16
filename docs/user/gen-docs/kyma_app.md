# kyma app

Manages applications on the Kubernetes cluster.

## Synopsis

Use this command to manage applications on the Kubernetes cluster.

```bash
kyma app <command> [flags]
```

## Available Commands

```text
  push - Push the application to the Kubernetes cluster
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma](kyma.md)                   - A simple set of commands to manage a Kyma cluster
* [kyma app push](kyma_app_push.md) - Push the application to the Kubernetes cluster
