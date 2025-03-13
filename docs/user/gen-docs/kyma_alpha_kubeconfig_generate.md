# kyma alpha kubeconfig generate

Generate kubeconfig with a Service Account-based or oidc tokens.

## Synopsis

Use this command to generate kubeconfig file with a Service Account-based or oidc tokens

```bash
kyma alpha kubeconfig generate [flags]
```

## Examples

```bash
# generate a kubeconfig with a ServiceAccount-based token and certificate
  kyma@v3 alpha kubeconfig generate --serviceaccount <sa_name> --clusterrole <cr_name> --namespace <ns_name> --permanent

# generate a kubeconfig with an OIDC token
  kyma@v3 alpha kubeconfig generate --token <token>

# generate a kubeconfig with an requested OIDC token
  kyma@v3 alpha kubeconfig generate --id-token-request-url <url>
```

## Flags

```text
      --clusterrole string      Name of the Cluster Role to bind the Service Account to
      --namespace string        Namespace in which the resource is created (default "default")
      --output string           Path to the kubeconfig file output. If not provided, the kubeconfig will be printed
      --permanent               Determines if the token is valid indefinitely
      --serviceaccount string   Name of the Service Account to be created
      --time string             Determines how long the token should be valid, by default 1h (use h for hours and d for days) (default "1h")
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
```

## See also

* [kyma alpha kubeconfig](kyma_alpha_kubeconfig.md) - Manages access to the Kyma cluster
