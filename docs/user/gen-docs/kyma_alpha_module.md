# kyma alpha module

Manages Kyma modules.

## Synopsis

Use this command to manage modules in the Kyma cluster.

```bash
kyma alpha module <command> [flags]
```

## Available Commands

```text
  add      - Add a module
  catalog  - Lists modules catalog
  delete   - Deletes a module
  list     - Lists the installed modules
  manage   - Sets the module to the managed state
  unmanage - Sets a module to the unmanaged state
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
```

## See also

* [kyma alpha](kyma_alpha.md)                                 - Groups command prototypes for which the API may still change
* [kyma alpha module add](kyma_alpha_module_add.md)           - Add a module
* [kyma alpha module catalog](kyma_alpha_module_catalog.md)   - Lists modules catalog
* [kyma alpha module delete](kyma_alpha_module_delete.md)     - Deletes a module
* [kyma alpha module list](kyma_alpha_module_list.md)         - Lists the installed modules
* [kyma alpha module manage](kyma_alpha_module_manage.md)     - Sets the module to the managed state
* [kyma alpha module unmanage](kyma_alpha_module_unmanage.md) - Sets a module to the unmanaged state
