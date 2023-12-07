---
title: kyma
---

Controls a Kyma cluster.

## Synopsis

Kyma is a flexible and easy way to connect and extend enterprise applications in a cloud-native world.
Kyma CLI allows you to install and manage Kyma.



```bash
kyma [flags]
```

## Flags

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner).
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma alpha](kyma_alpha.md)	 - Experimental commands
* [kyma apply](kyma_apply.md)	 - Applies local resources to the Kyma cluster.
* [kyma completion](kyma_completion.md)	 - Generates bash or zsh completion scripts.
* [kyma create](kyma_create.md)	 - Creates resources on the Kyma cluster.
* [kyma dashboard](kyma_dashboard.md)	 - Runs the Kyma dashboard locally and opens it directly in a web browser.
* [kyma get](kyma_get.md)	 - Gets Kyma-related resources.
* [kyma import](kyma_import.md)	 - Imports certificates to local certificates storage or adds domains to the local host file.
* [kyma init](kyma_init.md)	 - Creates local resources for your project.
* [kyma provision](kyma_provision.md)	 - Provisions a cluster for Kyma installation.
* [kyma run](kyma_run.md)	 - Runs resources.
* [kyma sync](kyma_sync.md)	 - Synchronizes the local resources for your Function.
* [kyma version](kyma_version.md)	 - Displays the version of Kyma CLI and of the connected Kyma cluster.

