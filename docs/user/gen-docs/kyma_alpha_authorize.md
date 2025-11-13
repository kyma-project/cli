# kyma alpha authorize

Authorize a subject (user, group, or service account) with Kyma RBAC resources.

## Synopsis

Create a RoleBinding or ClusterRoleBinding that grants access to a Kyma role or cluster role for a user, group, or service account.

```bash
kyma alpha authorize <authTarget> [flags]
```

## Available Commands

```text
  repository - Configure trust between a Kyma cluster and a GitHub repository
```

## Examples

```bash
  # Bind a user to a namespaced Role (RoleBinding)
  kyma alpha authorize user --name alice --role view --namespace dev

  # Bind multiple users to a namespaced Role (RoleBinding)
  kyma alpha authorize user --name alice,bob,james --role view --namespace dev

  # Bind a group cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize group --name team-observability --clusterrole kyma-read-all --cluster-wide

  # Bind a service account to a ClusterRole within a namespace (RoleBinding referencing a ClusterRole)
  kyma alpha authorize serviceaccount --name deployer-sa --clusterrole edit --namespace staging

  # Preview (dry-run) the YAML for a RoleBinding without applying
  kyma alpha authorize user --name bob --role operator --namespace ops --dry-run -o yaml

  # Generate JSON for a cluster-wide binding
  kyma alpha authorize user --name ci-bot --clusterrole kyma-admin --cluster-wide -o json
```

## Flags

```text
      --cluster-wide            Create a ClusterRoleBinding for cluster-wide access (requires --clusterrole)
      --clusterrole string      ClusterRole name to bind (for ClusterRoleBinding with --cluster-wide, or RoleBinding in namespace)
      --dry-run                 Preview the YAML/JSON output without applying resources to the cluster
      --name stringSlice        Name(s) of the subject(s) to authorize (required) (default "[]")
      --namespace string        Namespace for RoleBinding (required when binding a Role or binding a ClusterRole to a specific namespace)
  -o, --output string           Output format for dry-run (yaml or json)
      --role string             Role name to bind (creates RoleBinding in specified namespace)
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha](kyma_alpha.md)                                           - Groups command prototypes for which the API may still change
* [kyma alpha authorize repository](kyma_alpha_authorize_repository.md) - Configure trust between a Kyma cluster and a GitHub repository
