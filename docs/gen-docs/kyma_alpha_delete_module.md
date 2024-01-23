---
title: kyma alpha delete module
---

Deletes a module from the cluster or the given Kyma resource.

## Synopsis

Use this command to delete Kyma modules from the cluster.

### Detailed description

For more information on Kyma modules, see the 'create module' command.

This command disables an active module in the cluster.


```bash
kyma alpha delete module [name] [flags]
```

## Examples

```bash

Delete "my-module" from the "alpha" channel from "default-kyma" in "kyma-system" Namespace
		kyma alpha delete module my-module -c alpha -n kyma-system -k default-kyma

```

## Flags

```bash
  -c, --channel string     Module's channel to use.
  -f, --force-conflicts    Forces the patching of Kyma spec modules in case their managed field was edited by a source other than Kyma CLI.
  -k, --kyma-name string   Kyma resource to use. If empty, 'default-kyma' is used. (default "default-kyma")
  -n, --namespace string   Kyma Namespace to use. If empty, the default 'kyma-system' Namespace is used. (default "kyma-system")
  -t, --timeout duration   Maximum time for the operation to disable a module. (default 1m0s)
  -w, --wait               Wait until the given Kyma resource is ready.
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

* [kyma alpha delete](kyma_alpha_delete.md)	 - Deletes a resource from the Kyma cluster.

