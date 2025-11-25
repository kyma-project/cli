# kyma module add

Add a module.

## Synopsis

Use this command to add a module.

```bash
kyma module add <module> [flags]
```

## Examples

```bash
  # Add the Keda module with the default CR
  kyma module add keda --default-cr

  # Add the Keda module with a custom CR from a file
  kyma module add keda --cr-path ./keda-cr.yaml

  ## Add a community module with a default CR and auto-approve the SLA
  #  passed argument must be in the format <namespace>/<module-template-name>
  #  the module must be pulled from the catalog first using the 'kyma module pull' command
  kyma module add my-namespace/my-module-template-name --default-cr --auto-approve
```

## Flags

```text
      --auto-approve            Automatically approve community module installation
  -c, --channel string          Name of the Kyma channel to use for the module
      --cr-path string          Path to the custom resource file
      --default-cr              Deploys the module with the default CR
      --version string          Specifies version of the community module to install
      --context string          The name of the kubeconfig context to use
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
