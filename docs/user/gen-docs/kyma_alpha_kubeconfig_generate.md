# kyma alpha kubeconfig generate

Generate kubeconfig with a Service Account-based or oidc tokens.

## Synopsis

Use this command to generate kubeconfig file with a Service Account-based or oidc tokens

```bash
kyma alpha kubeconfig generate [flags]
```

## Examples

```bash
## Generate a permanent access (kubeconfig) for a new or existing ServiceAccount 
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --permanent

## Generate a permanent access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a namespaced binding to a given ClusterRole
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --clusterrole <cr_name> --permanent

## Generate a permanent access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a namespaced binding to a given Role
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --role <r_name> --permanent

## Generate time-constrained access (kubeconfig) for a new or existing ServiceAccount in a given namespace with a cluster-wide binding to a given ClusterRole
  kyma alpha kubeconfig generate --serviceaccount <sa_name> --namespace <ns_name> --clusterrole <cr_name> --cluster-wide --time 2h
  
## Generate a kubeconfig with an OIDC token
  kyma alpha kubeconfig generate --token <token>

## Generate a kubeconfig with an OIDC token based on a kubeconfig from the CIS
  kyma alpha kubeconfig generate --token <token> --credentials-path <cis_credentials>

## Generate a kubeconfig with an requested OIDC token with audience option
  export ACTIONS_ID_TOKEN_REQUEST_TOKEN=<token>
  kyma alpha kubeconfig generate --id-token-request-url <url> --audience <audience>

## Generate a kubeconfig with an requested OIDC token with url from env
  export ACTIONS_ID_TOKEN_REQUEST_URL=<url>
  export ACTIONS_ID_TOKEN_REQUEST_TOKEN=<token>
  kyma alpha kubeconfig generate
```

## Flags

```text
      --audience string                     Audience of the token
      --certificate-authority string        Path to a cert file for the certificate authority
      --certificate-authority-data string   Base64 encoded certificate authority data
      --cluster-wide                        Determines if the binding to the ClusterRole is cluster-wide
      --clusterrole string                  Name of the ClusterRole to bind the Service Account to (optional)
      --id-token-request-url string         URL to request the ID token, defaults to ACTIONS_ID_TOKEN_REQUEST_URL env variable
  -n, --namespace string                    Namespace in which the subject Service Account is to be found or is created (default "default")
      --oidc-name string                    Name of the OIDC custom resource from which the kubeconfig is generated
      --output string                       Path to the kubeconfig file output. If not provided, the kubeconfig is printed
      --permanent                           Determines if the token is valid indefinitely
      --role string                         Name of the Role in the given namespace to bind the Service Account to (optional)
  -s, --server string                       The address and port of the Kubernetes API server
      --serviceaccount string               Name of the Service Account (in the given namespace) to be used as a subject of the generated kubeconfig. If the Service Account does not exist, it is created
      --time string                         Determines how long the token is valid, by default 1h (use h for hours and d for days) (default "1h")
      --token string                        Token used in the kubeconfig
      --context string                      The name of the kubeconfig context to use
  -h, --help                                Help for the command
      --kubeconfig string                   Path to the Kyma kubeconfig file
      --show-extensions-error               Prints a possible error when fetching extensions fails
      --skip-extensions                     Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha kubeconfig](kyma_alpha_kubeconfig.md) - Manages access to the Kyma cluster
