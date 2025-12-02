# kyma alpha module

Manages Kyma modules.

## Synopsis

Use this command to manage modules in the Kyma cluster.

```bash
kyma alpha module <command> [flags]
```

## Available Commands

```text
  catalog - Lists modules catalog
```

## Flags

```text
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha](kyma_alpha.md)                               - Groups command prototypes for which the API may still change
* [kyma alpha module catalog](kyma_alpha_module_catalog.md) - Lists modules catalog
