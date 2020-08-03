---
title: Use Kyma CLI
type: Details
---

Kyma CLI comes with a set of commands, each of which has its own specific set of flags. Use them to provision the cluster locally or using a chosen cloud provider, install, and test Kyma.

For the commands and flags to work, they need to follow this syntax:

```bash
kyma {COMMAND} {FLAGS}
```

- **{COMMAND}** specifies the operation you want to perform, such as provisioning the cluster or installing Kyma.
- **{FLAGS}** specifies optional flags you can use to enrich your command.

See the example:

```bash
kyma install -s latest
```

>**TIP:** Documentation for particular commands is generated automatically with the code. See [the full list of commands and flags](https://github.com/kyma-project/cli/tree/master/docs/gen-docs).

|     Command        | Child commands   |  Description  | Example |
|--------------------|----------------|---------------|---------|
| [`completion`](/docs/gen-docs/kyma_completion.md)| None| Generates and displays the bash or zsh completion script. | `kyma completion`|
| [`console`](/docs/gen-docs/kyma_console.md)| None| Launches Kyma Console in a browser window. | `kyma console` |
| [`install`](/docs/gen-docs/kyma_install.md)| None| Installs Kyma on a cluster based on the current or specified release. | `kyma install`|
| [`provision`](/docs/gen-docs/kyma_provision.md)| [`minikube`](/docs/gen-docs/kyma_provision_minikube.md)<br> [`gardener`](/docs/gen-docs/kyma_provision_gardener.md) <br> [`gke`](/docs/gen-docs/kyma_provision_gke.md) <br> [`aks`](/docs/gen-docs/kyma_provision_aks.md)| Provisions a new cluster on a platform of your choice. Currently, this command supports cluster provisioning on GCP, Azure, Gardener, and Minikube. | `kyma provision minikube`|
| [`test`](/docs/gen-docs/kyma_test.md)|[`definitions`](/docs/gen-docs/kyma_test_definitions.md)<br> [`delete`](/docs/gen-docs/kyma_test_delete.md) <br> [`list`](/docs/gen-docs/kyma_test_list.md) <br> [`run`](/docs/gen-docs/kyma_test_run.md) <br> [`status`](/docs/gen-docs/kyma_test_status.md)<br> [`logs`](/docs/gen-docs/kyma_test_logs.md) <br> | Runs and manages tests on a provisioned Kyma cluster. Using child commands, you can run tests, view test definitions, list and delete test suites, display test status, and fetch the logs of the tests.| `kyma test run` |
| [`version`](/docs/gen-docs/kyma_version.md)|None| Shows the cluster version and the Kyma CLI version.| `kyma version` |
