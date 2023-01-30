---
title: kyma alpha sign module
---

Signs all module resources from an unsigned component descriptor that's hosted in a remote OCI registry

## Synopsis

Use this command to sign a Kyma module.

### Detailed description

This command signs all module resources recursively based on an unsigned component descriptor hosted in an OCI registry with the provided private key. Then, the output (component-descriptor.yaml) is saved in the descriptor path (default: ./mod) as a signed component descriptor. If signed-registry are provided, the command pushes the signed component descriptor.


```bash
kyma alpha sign module MODULE_NAME MODULE_VERSION [flags]
```

## Flags

```bash
  -c, --credentials string       Basic authentication credentials for the given registry in the format user:password
      --insecure                 Use an insecure connection to access the registry.
      --mod-path string          Specifies the path where the signed component descriptor will be stored (default "./mod")
      --nameMapping string       Overrides the OCM Component Name Mapping, one of: "urlPath" or "sha256-digest" (default "urlPath")
      --private-key string       Specifies the path where the private key used for signing
      --registry string          Repository context url where unsigned component descriptor located
      --signature-name string    name of the signature for signing
      --signed-registry string   Repository context url where signed component descriptor located
  -t, --token string             Authentication token for the given registry (alternative to basic authentication).
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

* [kyma alpha sign](kyma_alpha_sign.md)	 - Signs all module resources from an unsigned component descriptor that's hosted in a remote OCI registry

