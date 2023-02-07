---
title: kyma alpha disable module
---

Disables a module in the cluster or in the given Kyma resource

## Synopsis

Use this command to disable Kyma modules active in the cluster.

### Detailed description

For more information on Kyma modules, see the 'create module' command.

This command disables an active module in the cluster.


```bash
kyma alpha disable module [name] [flags]
```

## Examples

```bash
Examples:
TODO

```

## Flags

```bash
  -c, --channel string     The channel of the module to use.
  -k, --kyma-name string   The name of the Kyma to use. An empty name uses 'default-kyma' (default "default-kyma")
  -n, --namespace string   The namespace of the Kyma to use. An empty namespace uses 'default' (default "default")
  -t, --timeout duration   Maximum time for the operation to disable a module. (default 1m0s)
  -w, --wait               Wait until the given Kyma resource is ready
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

* [kyma alpha disable](kyma_alpha_disable.md)	 - Disables a resource on the Kyma cluster.

