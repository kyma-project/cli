# kyma module delete

Deletes a module.

## Synopsis

Use this command to delete a module.

```bash
kyma module delete <module> [flags]
```

## Flags

```text
      --auto-approve            Automatically approves module removal
      --community               Delete the community module (if set, the operation targets a community module instead of a core module)
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
  -q, --quiet                   Suppress non-essential output
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
