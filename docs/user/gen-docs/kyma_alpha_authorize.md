# kyma alpha authorize

Configure trust between a Kyma cluster and a GitHub repository.

## Synopsis

Configure trust between a Kyma cluster and a GitHub repository by creating an OpenIDConnect resource and RoleBinding or ClusterRoleBinding

```bash
kyma alpha authorize repository [flags]
```

## Flags

```text
      --client-id string        OIDC client ID (audience) expected in the token (required)
      --cluster-wide            If true, create a ClusterRoleBinding; otherwise, a RoleBinding
      --clusterrole string      ClusterRole name to bind (usable for RoleBinding or ClusterRoleBinding)
      --dry-run                 Print resources without applying
      --issuer-url string       OIDC issuer (default "https://token.actions.githubusercontent.com")
      --name string             Name for the OpenIDConnect resource (optional; default derives from clientId)
      --namespace string        Namespace for RoleBinding (required if not cluster-wide and binding a Role or namespaced ClusterRole)
  -o, --output string           Output format (yaml or json)
      --prefix string           Username prefix for the repository claim (e.g., gh-oidc:)
      --repository string       GitHub repo in owner/name format (e.g., kyma-project/cli) (required)
      --role string             Role name to bind (namespaced)
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha](kyma_alpha.md) - Groups command prototypes for which the API may still change
