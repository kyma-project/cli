# kyma alpha module catalog

Lists modules catalog.

## Synopsis

Use this command to list all available Kyma modules.

```bash
kyma alpha module catalog [flags]
```

## Examples

```bash
  # List all modules available in the cluster (core and community)
  kyma module catalog

  # List available community modules from the official repository
  kyma module catalog --remote

  # List available community modules from a specific remote URL
  kyma module catalog --remote=https://example.com/modules.json

  # List available community modules from multiple remote URLs
  kyma module catalog --remote=https://example.com/modules1.json,https://example.com/modules2.json

  # Output catalog as JSON
  kyma module catalog -o json

  # List remote community modules in YAML format
  kyma module catalog --remote -o yaml
```

## Flags

```text
  -o, --output string             Output format (Possible values: table, json, yaml)
      --remote bool or []string   Fetch modules from the official repository or specify custom URL(s) (default "false")
      --context string            The name of the kubeconfig context to use
  -h, --help                      Help for the command
      --kubeconfig string         Path to the Kyma kubeconfig file
      --show-extensions-error     Prints a possible error when fetching extensions fails
      --skip-extensions           Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha module](kyma_alpha_module.md) - Manages Kyma modules
