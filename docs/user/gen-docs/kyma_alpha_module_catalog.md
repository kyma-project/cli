# kyma alpha module catalog

Lists modules catalog.

## Synopsis

Use this command to list all available Kyma modules.

```bash
kyma alpha module catalog [flags]
```

## Examples

```bash

  # List all available modules from all origins
  kyma module catalog

  # List only official Kyma modules managed by KLM with SLA
  kyma module catalog --origin kyma

  # List only community modules (not officially supported)
  kyma module catalog --origin community

  # List only community modules already available on the cluster
  kyma module catalog --origin cluster

  # List modules from multiple origins
  kyma module catalog --origin kyma,community

  # Output catalog as JSON
  kyma module catalog -o json

  # List official Kyma modules in YAML format
  kyma module catalog --origin kyma -o yaml
```

## Flags

```text
      --origin stringSlice      Specifies the source of the module (default "[kyma,community,cluster]")
  -o, --output string           Output format (Possible values: table, json, yaml)
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma alpha module](kyma_alpha_module.md) - Manages Kyma modules
