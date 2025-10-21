# kyma module pull

Pull a module from a remote repository.

## Synopsis

Pull a module from a remote repository to make it available for installation on the cluster.

This command downloads module templates and resources from remote repositories,
making them available locally for subsequent installation. Community modules
must be pulled before they can be installed using the 'kyma module add' command.

Examples:
  # Pull a specific community module
  kyma module pull `{YOUR_COMMUNITY_MODULE}`

  # Pull a module with a specific version into specific namespace
  kyma module pull `{YOUR_COMMUNITY_MODULE}` --version `{YOUR_MODULE_VERSION}` --namespace `{YOUR_MODULE_NAMESPACE}`

```bash
kyma module pull <module-name> [flags]
```

## Flags

```text
  -n, --namespace string        Destination namespace where the module is stored (default "default")
  -v, --version string          Specify version of the community module to pull
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
