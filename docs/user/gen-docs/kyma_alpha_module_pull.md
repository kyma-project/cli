# kyma alpha module pull

Pulls a module from a remote repository.

## Synopsis

Pulls a module from a remote repository to make it available for installation in the cluster.

This command downloads module templates and resources from remote repositories,
making them available locally for subsequent installation. Community modules
must be pulled before they can be installed using the 'kyma module add' command.

```bash
kyma alpha module pull <module-name> [flags]
```

## Examples

```bash
  # Pull a specific community module
  kyma alpha module pull community-module-name

  # Pull the latest version of a module into specific namespace
  kyma alpha module pull community-module-name --namespace module-namespace

  # Pull a module with a specific version into specific namespace
  kyma alpha module pull community-module-name --version v1.0.0 --namespace module-namespace

  # Pull a module from a custom remote repository URL
  kyma alpha module pull community-module-name --remote-url https://example.com/modules.json
```

## Flags

```text
      --force                   Forces application of the module template, overwriting if it already exists
  -n, --namespace string        Destination namespace where the module is stored (default "default")
      --remote-url string       Specifies the community modules repository URL (defaults to official repository)
  -v, --version string          Specifies the version of the community module to pull
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha module](kyma_alpha_module.md) - Manages Kyma modules
