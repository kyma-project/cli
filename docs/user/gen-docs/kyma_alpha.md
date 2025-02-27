# kyma alpha

Groups command prototypes for which the API may still change.

## Synopsis

A set of alpha prototypes that may still change. Use them in automation at your own risk.

```bash
kyma alpha <command> [flags]
```

## Available Commands

```text
  access             - Produces a kubeconfig with a Service Account-based token and certificate
  app                - Manages applications on the Kubernetes cluster
  hana               - Manages an SAP HANA instance in the Kyma cluster
  help               - Help about any command
  module             - Manages Kyma modules
  oidc               - Creates kubeconfig with an OIDC token
  provision          - Provisions a Kyma cluster on SAP BTP
  reference-instance - Adds an instance reference to a shared service instance
```

## Flags

```text
      --kubeconfig string       Path to the Kyma kubeconfig file
  -h, --help                    Help for the command
      --show-extensions-error   Prints a possible error when fetching extensions fails
```

## See also

* [kyma](kyma.md)                                                   - A simple set of commands to manage a Kyma cluster
* [kyma alpha access](kyma_alpha_access.md)                         - Produces a kubeconfig with a Service Account-based token and certificate
* [kyma alpha app](kyma_alpha_app.md)                               - Manages applications on the Kubernetes cluster
* [kyma alpha hana](kyma_alpha_hana.md)                             - Manages an SAP HANA instance in the Kyma cluster
* [kyma alpha module](kyma_alpha_module.md)                         - Manages Kyma modules
* [kyma alpha oidc](kyma_alpha_oidc.md)                             - Creates kubeconfig with an OIDC token
* [kyma alpha provision](kyma_alpha_provision.md)                   - Provisions a Kyma cluster on SAP BTP
* [kyma alpha reference-instance](kyma_alpha_reference-instance.md) - Adds an instance reference to a shared service instance
