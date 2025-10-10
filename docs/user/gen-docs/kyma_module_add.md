# kyma module add

Add a module.

## Synopsis

Use this command to add a module.

```bash
kyma module add <module> [flags]
```

## Flags

```text
      --auto-approve            Automatically approve community module installation
  -c, --channel string          Name of the Kyma channel to use for the module
      --community               Install a community module (no official support, no binding SLA)
      --cr-path string          Path to the custom resource file
      --default-cr              Deploys the module with the default CR
      --origin string           Specifies the source of the module (kyma or custom name)
      --version string          Specifies version of the community module to install
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the target Kyma environment
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
