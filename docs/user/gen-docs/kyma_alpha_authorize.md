# kyma alpha authorize

Authorizes a subject (user, group, or service account) with Kyma RBAC resources.

## Synopsis

Create a RoleBinding or ClusterRoleBinding that grants access to a Kyma role or cluster role for a user, group, or service account.

```bash
kyma alpha authorize
```

## Available Commands

```text
  group          - Authorizes a group with Kyma RBAC resources
  repository     - Configures a trust between a Kyma cluster and a GitHub repository
  serviceaccount - Authorizes a service account with Kyma RBAC resources
  user           - Authorizes a user with Kyma RBAC resources
```

## Flags

```text
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha](kyma_alpha.md)                                                   - Groups command prototypes for which the API may still change
* [kyma alpha authorize group](kyma_alpha_authorize_group.md)                   - Authorizes a group with Kyma RBAC resources
* [kyma alpha authorize repository](kyma_alpha_authorize_repository.md)         - Configures a trust between a Kyma cluster and a GitHub repository
* [kyma alpha authorize serviceaccount](kyma_alpha_authorize_serviceaccount.md) - Authorizes a service account with Kyma RBAC resources
* [kyma alpha authorize user](kyma_alpha_authorize_user.md)                     - Authorizes a user with Kyma RBAC resources
