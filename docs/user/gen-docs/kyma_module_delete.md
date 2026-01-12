# kyma module delete

Deletes a module.

## Synopsis

Use this command to delete a module.

```bash
kyma module delete <module> [flags]
```

## Examples

```bash
  # Delete the Keda module
  kyma module delete keda

  ## Delete a community module and auto-approve the deletion
  #  passed argument must be in the format <namespace>/<module-template-name>
  #  the format of the passed argument can be read from the 'kyma module catalog' command from the 'origin' column
  kyma module delete my-namespace/my-community-module-1.0.0 --auto-approve
```

## Flags

```text
      --auto-approve            Automatically approves module removal
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
