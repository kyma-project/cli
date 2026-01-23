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
  kyma alpha module catalog

  # List available community modules from the official repository
  kyma alpha module catalog --remote

  # List available community modules from a specific remote URL
  kyma alpha module catalog --remote-url=https://example.com/modules.json

  # List available community modules from multiple remote URLs
  kyma alpha module catalog --remote=https://example.com/modules1.json,https://example.com/modules2.json

  # Output catalog as JSON
  kyma alpha module catalog -o json

  # List remote community modules in YAML format
  kyma alpha module catalog --remote -o yaml
```

## Flags

```text
  -o, --output string            Output format (Possible values: table, json, yaml)
      --remote                   Fetch modules from the official repository
      --remote-url stringSlice   List of URLs to custom community module repositories (default "[]")
      --context string           The name of the kubeconfig context to use
  -h, --help                     Help for the command
      --kubeconfig string        Path to the Kyma kubeconfig file
      --show-extensions-error    Prints a possible error when fetching extensions fails
      --skip-extensions          Skips fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha module](kyma_alpha_module.md) - Manages Kyma modules
