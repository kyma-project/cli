---
title: kyma alpha enable module
---

Enables a module in the cluster or in the given Kyma resource

## Synopsis

Use this command to enable Kyma modules available in the cluster.

### Detailed description

For more information on Kyma modules, see the 'create module' command.

This command enables an available module in the cluster. 
A module is available when a ModuleTemplate is found for instantiating it with proper defaults.


```bash
kyma alpha enable module [name] [flags]
```

## Examples

```bash
Example:
Enable "my-module" from "alpha"" channel in "default-kyma" from "kyma-system" namespace
		kyma alpha enable module my-module -c alpha -n kyma-system -k default-kyma

```

## Flags

```bash
  -c, --channel string     The channel of the module to enable.
  -k, --kyma-name string   The name of the Kyma to use. An empty name uses 'default-kyma' (default "default-kyma")
  -n, --namespace string   The namespace of the Kyma to use. An empty namespace defaults to 'kyma-system' (default "kyma-system")
  -t, --timeout duration   Maximum time for the operation to enable a module. (default 1m0s)
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

* [kyma alpha enable](kyma_alpha_enable.md)	 - Enables a resource on the Kyma cluster.

