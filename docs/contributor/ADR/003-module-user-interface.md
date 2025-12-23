# Output Error Format

Creation date: 2025.10.11

## Introduction

In this ADR, I propose a new API for the most basic functionalities in the `kyma module` commands group (mostly around handling community modules). This ADR is built based on the following [issue](https://github.com/kyma-project/cli/issues/2765) and its goal is to keep all important information in one place.

## Description

In the latest release (`3.2.0`), operations on community modules are not intuitive. For example, to delete a module, the user must first list the installed modules, then print the catalog to get the origin of the module, and then use this origin to remove the community module. The flow:

```bash
$ kyma module list

NAME              VERSION   CR POLICY   MANAGED   MODULE STATUS   INSTALLATION STATUS
...
cap-operator      0.20.1    N/A         false     NotRunning      Ready

$ kyma module catalog

NAME           AVAILABLE VERSIONS   ORIGIN
...
cap-operator   0.20.1               default/cap-operator-0.20.1

$ kyma module delete default/cap-operator-0.20.1
...
```

The proposition of changes:

- To simplify the view and build similarities to the `Busola`, we should split the tables (`module list` and `module catalog`) into two tables, first for core modules, and the second one for community modules.

    ```bash
    $ kyma module list
    
    MODULE         VERSION           CR POLICY         MANAGED   MODULE STATUS   INSTALLATION STATUS
    api-gateway    3.3.0(regular)    CreateAndDelete   true      Ready           Ready
    btp-operator   1.2.20(regular)   CreateAndDelete   true      Ready           Ready
    eventing       1.4.0(fast)       CreateAndDelete   true      Ready           Ready
    istio          1.23.1(regular)   CreateAndDelete   true      Ready           Ready
    keda           1.9.0(fast)       CreateAndDelete   true      Ready           Ready
    nats           1.2.2(fast)       CreateAndDelete   true      Ready           Ready
    serverless     1.9.1(fast)       Ignore            false     Ready           Unmanaged
    
    COMMUNITY MODULE          VERSION   MODULE STATUS   INSTALLATION STATUS
    default/docker-registry   0.10.0    NotRunning      Unknown
    ```

    ```bash
    $ kyma module catalog
    
    MODULE                  AVAILABLE VERSIONS
    api-gateway             3.3.0(regular), 3.4.1(fast)
    application-connector   1.1.17(fast), 1.1.17-experimental(experimental)
    btp-operator            1.2.20(regular), 1.2.22(fast)
    cap-operator            0.21.0(fast)
    cfapi                   0.3.1(experimental)
    cloud-manager           1.6.1(experimental)
    connectivity-proxy      1.1.2(fast), 1.1.2-experimental(experimental)
    docker-registry         0.10.0(experimental)
    eventing                1.4.0(fast)
    istio                   1.23.1(regular), 1.23.2(fast), 1.23.2-experimental(experimental)
    keda                    1.8.2(regular), 1.9.0(fast)
    nats                    1.2.2(experimental)
    serverless              1.8.4(regular), 1.9.1(experimental)
    telemetry               1.52.0(regular), 1.53.0(fast), 1.53.0-experimental(experimental)
    transparent-proxy       1.9.0(fast)
    ztis-agent              0.20.0-experimental(experimental)
    
    COMMUNITY MODULE          AVAILABLE VERSIONS
    default/cap-operator      0.20.1
    default/docker-registry   0.10.0
    default/registry-proxy    0.14.0
    ```

- Remote catalog should not be a part of the default output of the `module catalog` command. Instead, to display modules available to pull, the user should add the `--remote` or `--remote=<URL>` flag.

    ```bash
    $ kyma module catalog --remote
    
    MODULE            AVAILABLE VERSIONS
    cap-operator      0.20.1
    docker-registry   0.10.0
    registry-proxy    0.14.0
    ```

- `--remote` should be implemented also in `pull` command.

- The `ORIGIN` column should not be present in any table because it's confusing and not helpful.

- The names for installed or available (on cluster) community modules should be simplified to format `<MODULE_TEMPLATE_NAMESPACE>/<MODULE_NAME>` instead of `< MODULE_TEMPLATE_NAMESPACE>/< MODULE_TEMPLATE_NAME>`.

- To support automations, we should allow using the `--auto-approve` that will accept SLA and choose the latest available release.

    ```bash
    $ kyma module add default/docker-registry --auto-approve
    
    Warning:
      You are about to install a community module.
      Community modules are not officially supported and come with no binding Service Level Agreement (SLA).
    
    A few community module versions are available on the cluster: [0.9.0, 0.10.0]. Choosing the latest: 0.10.0.
    
    The community module docker-registry in version 0.10.0 is installed
    ```

- Different output formats (other than the table) should combine modules and community modules into one object to allow external tools like `jq` or `yq` to go through this output:

    ```bash
    $ kyma module catalog -ojson
    
    {
      "modules": [
        ...
      ],
      "communityModules": [
        ...
      ]
    }
    ```

## Example flow

describe catalog->pull->add->pull new->upgrade->delete flow

```bash
# FInd available module in remote catalog
$ kyma module catalog --remote

MODULE            AVAILABLE VERSIONS
cap-operator      0.20.1
docker-registry   0.10.0
registry-proxy    0.14.0

# Pull module
$ kyma module pull docker-registry --namespace default

...

# Add module
$ kyma module add default/docker-registry --auto-approve

Warning:
  You are about to install a community module.
  Community modules are not officially supported and come with no binding Service Level Agreement (SLA).

The community module docker-registry in version 0.9.0 is installed

##########################
# some time later when new version of the community docker-registry is out

# Pull new module version
$ kyma module pull docker-registry --namespace default

...

# Install new version
$ kyma module add default/docker-registry --auto-approve

Warning:
  You are about to install a community module.
  Community modules are not officially supported and come with no binding Service Level Agreement (SLA).

A few community module versions are available on the cluster: [0.9.0, 0.10.0]. Choosing the latest: 0.10.0.

The community module docker-registry in version 0.10.0 is installed

# Remove community module
$ kyma module delete default/docker-registry

...
```

## Reasons

The module API should be easy to manage and consistent from the user's perspective. It should allow users to manage clusters in the CI/CD system and on a local machine. To do so, we've collected all suggestions from previous tasks, like:

- [Issue 2734](https://github.com/kyma-project/cli/issues/2734#issuecomment-3593132479)
- [Issue 2675](https://github.com/kyma-project/cli/issues/2675#issuecomment-3499464158)
- [Issue 2670](https://github.com/kyma-project/cli/issues/2670#issuecomment-3487979099)

The main goal is to allow managing the community modules in a more generic way and separate this flow from the core modules' flow. It should be easy to figure out which module is in the local catalog (on a cluster) and which one is available on the remote and should be pulled first. Additionally, it should be intuitive for users to know which module is a community and which is the core one.
