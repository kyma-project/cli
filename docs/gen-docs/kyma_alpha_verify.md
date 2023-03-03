---
title: kyma alpha verify
---

Verifies all module resources from a signed module component descriptor that's hosted in a remote OCI registry.

## Synopsis

Use this command to verify all module resources from a signed module descriptor that's hosted in a remote OCI registry.

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner).
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma alpha](kyma_alpha.md)	 - Experimental commands
* [kyma alpha verify module](kyma_alpha_verify_module.md)	 - Verifies the signature of a Kyma module bundled as an OCI container image.

