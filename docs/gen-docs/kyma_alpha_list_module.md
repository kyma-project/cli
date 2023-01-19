---
title: kyma alpha list module
---

list all available modules available for creation in the cluster

## Synopsis

Use this command to list Kyma modules availabel in the cluster.

### Detailed description

For more information on Kyma modules, see the "create module" command.

This command lists all available modules in the cluster. 
A module is available when a ModuleTemplate is found for instantiating it with proper defaults.

Optionally, you can manually add a release channel to filter available modules only for the given channel.

Also, you can specify a Kyma to use to lookup only the active modules within that Kyma instance. If this is specified,
the ModuleTemplates will also have a Field called "State" which will reflect the actual state of the module.

Finally, you can restrict and select a custom namespace for the command.


```bash
kyma alpha list module [flags]
```

## Examples

```bash
Examples:
List all modules
		kyma alpha list module
List all modules in the "regular" channel
		kyma alpha list module --channel regular
List all modules for the kyma "some-kyma" in the namespace "custom" in the "alpha" channel
		kyma alpha list module --kyma some-kyma -c alpha -n "custom"
List all modules for the kyma "some-kyma" in the "alpha" channel
		kyma alpha list module --kyma some-kyma -c alpha

```

## Flags

```bash
  -c, --channel string     Channel to use for the module template.
  -k, --kyma string        The namespaced name of the kyma to use to list active module templates in the form 'namespace/name'.
  -n, --namespace string   The namespace to use. An empty namespace uses 'default'
  -t, --timeout duration   Maximum time for the list operation to retrieve ModuleTemplates. (default 1m0s)
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma alpha list](kyma_alpha_list.md)	 - Lists resources on the Kyma cluster.

