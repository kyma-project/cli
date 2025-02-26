# kyma alpha access

Produces a kubeconfig with a Service Account-based token and certificate.

## Synopsis

Use this command to produce a kubeconfig with a Service Account-based token and certificate that is valid for a specified time or indefinitely.

```bash
kyma alpha access [flags]
```

## Flags

```text
      --clusterrole string      Name of the cluster role to bind the Service Account to
      --name string             Name of the Service Account to be created
      --namespace string        Namespace in which the resource is created (default "default")
      --output string           Path to the kubeconfig file output. If not provided, the kubeconfig will be printed
      --permanent               Determines if the token is valid indefinitely
      --time string             Determines how long the token should be valid, by default 1h (use h for hours and d for days) (default "1h")
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
```

## See also

* [kyma alpha](kyma_alpha.md) - Groups command prototypes for which the API may still change
