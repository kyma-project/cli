# kyma alpha authorize group

Authorize a Group with Kyma RBAC resources.

## Synopsis

Create a RoleBinding or ClusterRoleBinding that grants access to a Kyma role or cluster role for a Group.

```bash
kyma alpha authorize group [flags]
```

## Examples

```bash
  # Bind a group cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize group --name team-observability --clusterrole kyma-read-all --cluster-wide

  # Bind a group to a namespaced Role (RoleBinding)
  kyma alpha authorize group --name developers --role edit --namespace dev

  # Generate JSON for a cluster-wide binding
  kyma alpha authorize group --name ops-team --clusterrole cluster-admin --cluster-wide -o json
  
  # Preview (dry-run) the YAML for a RoleBinding without applying
  kyma alpha authorize group --name ops-team --role edit --namespace dev --dry-run -o yaml
```

## Flags

```text
      --binding-name string     Custom name for the RoleBinding or ClusterRoleBinding. If not specified, a name is auto-generated based on the role and subject
      --cluster-wide            Create a ClusterRoleBinding for cluster-wide access (requires --clusterrole)
      --clusterrole string      ClusterRole name to bind (for ClusterRoleBinding with --cluster-wide, or RoleBinding in namespace)
      --dry-run                 Preview the YAML/JSON output without applying resources to the cluster
      --force                   Forces application of the binding, overwriting if it already exists
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

* [kyma alpha authorize](kyma_alpha_authorize.md) - Authorize a subject (user, group, or service account) with Kyma RBAC resources
