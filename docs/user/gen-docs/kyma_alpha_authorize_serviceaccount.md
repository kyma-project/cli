# kyma alpha authorize serviceaccount

Authorize a ServiceAccount with Kyma RBAC resources.

## Synopsis

Create a RoleBinding or ClusterRoleBinding that grants access to a Kyma role or cluster role for a ServiceAccount.

```bash
kyma alpha authorize serviceaccount [flags]
```

## Examples

```bash
  # Bind a service account to a ClusterRole within a namespace (RoleBinding referencing a ClusterRole)
  kyma alpha authorize serviceaccount --name deployer-sa --clusterrole edit --namespace staging

  # Bind a service account cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize serviceaccount --name system-bot --clusterrole cluster-admin --cluster-wide

  # Specify a different namespace for the service account subject
  kyma alpha authorize serviceaccount --name remote-sa --sa-namespace tools --clusterrole view --namespace dev
   
  # Preview (dry-run) the YAML for a RoleBinding without applying
  kyma alpha authorize serviceaccount --name remote-sa --role edit --namespace dev --dry-run -o yaml 
  
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
      --sa-namespace string     Namespace for the service account subject. Defaults to the RoleBinding namespace when not specified.
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha authorize](kyma_alpha_authorize.md) - Authorize a subject (user, group, or service account) with Kyma RBAC resources
