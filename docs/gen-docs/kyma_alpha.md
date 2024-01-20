---
title: kyma alpha
---

Experimental commands

## Synopsis

Alpha commands are experimental, unreleased features that should only be used by the Kyma team. Use at your own risk.


## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner).
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](kyma.md)	 - Controls a Kyma cluster.
* [kyma alpha add](kyma_alpha_add.md)	 - Enables a resource in the Kyma cluster.
* [kyma alpha create](kyma_alpha_create.md)	 - Creates resources on the Kyma cluster.
* [kyma alpha delete](kyma_alpha_delete.md)	 - Disables a resource in the Kyma cluster.
* [kyma alpha deploy](kyma_alpha_deploy.md)	 - Deploys Kyma on a running Kubernetes cluster.
* [kyma alpha list](kyma_alpha_list.md)	 - Lists resources on the Kyma cluster.
* [kyma alpha sign](kyma_alpha_sign.md)	 - Signs all module resources from an unsigned module component descriptor that's hosted in a remote OCI registry
* [kyma alpha verify](kyma_alpha_verify.md)	 - Verifies all module resources from a signed module component descriptor that's hosted in a remote OCI registry

