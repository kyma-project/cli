# kyma alpha authorize repository

Configures a trust between a Kyma cluster and a GitHub repository.

## Synopsis

Configures a trust between a Kyma cluster and a GitHub repository by creating an OpenIDConnect resource and RoleBinding or ClusterRoleBinding

```bash
kyma alpha authorize repository [flags]
```

## Examples

```bash
  # Authorize a repository with a namespaced Role (RoleBinding)
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --role view --namespace dev

  # Authorize a repository cluster-wide to a ClusterRole (ClusterRoleBinding)
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --clusterrole kyma-read-all --cluster-wide

  # Bind a repository to a ClusterRole within a single namespace (RoleBinding referencing ClusterRole)
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --clusterrole edit --namespace staging

  # Preview (dry-run) the YAML without applying
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --role operator --namespace ops --dry-run -o yaml

  # Provide a custom OpenIDConnect resource name and username prefix
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --clusterrole kyma-admin --cluster-wide --name custom-oidc --prefix gh-oidc:

  # Add additional required claims to the OIDC resource
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --role view --namespace dev --required-claim environment=dev --required-claim workflow=main

  # Use JSON output to inspect resources before apply
  kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client --role view --namespace dev --dry-run -o json
```

## Flags

```text
      --client-id string                OIDC client ID (audience) expected in the token (required)
      --cluster-wide                    If true, creates a ClusterRoleBinding; otherwise, a RoleBinding
      --clusterrole string              ClusterRole name to bind (usable for RoleBinding and ClusterRoleBinding)
      --dry-run                         Prints resources without applying
      --issuer-url string               OIDC issuer (default "https://token.actions.githubusercontent.com")
      --name string                     Name for the OpenIDConnect resource (optional; default derives from clientId)
      --namespace string                Namespace where the RoleBinding is created (required unless --cluster-wide is set)
  -o, --output string                   Output format (yaml or json)
      --prefix string                   Username prefix for the repository claim (e.g., gh-oidc:)
      --repository string               GitHub repo in owner/name format (e.g., kyma-project/cli) (required)
      --required-claim stringToString   Additional required claims (key=value) for the OpenIDConnect resource (default "[]")
      --role string                     Role name to bind (namespaced)
      --context string                  The name of the kubeconfig context to use
  -h, --help                            Help for the command
      --kubeconfig string               Path to the Kyma kubeconfig file
      --show-extensions-error           Prints a possible error when fetching extensions fails
      --skip-extensions                 Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha authorize](kyma_alpha_authorize.md) - Authorizes a subject (user, group, or service account) with Kyma RBAC resources
