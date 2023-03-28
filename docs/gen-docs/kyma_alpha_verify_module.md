---
title: kyma alpha verify module
---

Verifies the signature of a Kyma module bundled as an OCI container image.

## Synopsis

Use this command to verify a Kyma module.

### Detailed description

Kyma modules can be cryptographically signed to ensure they are correct and distributed by a trusted authority. This command verifies the authenticity of a given module.


```bash
kyma alpha verify module --name MODULE_NAME --version MODULE_VERSION --registry MODULE_REGISTRY [flags]
```

## Flags

```bash
  -c, --credentials string    Basic authentication credentials for the given registry in the user:password format
      --insecure              Uses an insecure connection to access the registry.
      --key string            Specifies the path where a public key is used for signing.
      --name string           Name of the module.
      --name-mapping string   Overrides the OCM Component Name Mapping, Use: "urlPath" or "sha256-digest". (default "urlPath")
      --registry string       Context URL of the repository for the module. The repository's URL is automatically added to the repository's contexts in the module.
  -t, --token string          Authentication token for the given registry (alternative to basic authentication).
      --version string        Version of the module.
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner).
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma alpha verify](kyma_alpha_verify.md)	 - Verifies all module resources from a signed module component descriptor that's hosted in a remote OCI registry

