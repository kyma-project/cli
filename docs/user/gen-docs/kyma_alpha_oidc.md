# kyma alpha oidc

Creates kubeconfig with an OIDC token.

## Synopsis

Use this command to create kubeconfig with an OIDC token generated with a GitHub Actions token.

```bash
kyma alpha oidc [flags]
```

## Flags

```text
      --audience string               Audience of the token
      --credentials-path string       Path to the CIS credentials file
      --id-token-request-url string   URL to request the ID token, defaults to ACTIONS_ID_TOKEN_REQUEST_URL env variable
      --output string                 Path to the output kubeconfig file
      --token string                  Token used in the kubeconfig
  -h, --help                          Help for the command
      --kubeconfig string             Path to the Kyma kubeconfig file
      --show-extensions-error         Prints a possible error when fetching extensions fails
```

## See also

* [kyma alpha](kyma_alpha.md) - Groups command prototypes for which the API may still change
