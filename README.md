# Kyma CLI

## Overview

Kyma CLI is a command line tool which supports [Kyma](https://github.com/kyma-project/kyma) developers. It provides a set of commands you can use to install, manage, and test Kyma.

## Prerequisites

Kyma CLI requires the following software:

- [kubectl](https://github.com/kubernetes/kubectl) 
- [Minikube](https://github.com/kubernetes/minikube)

## Installation

Use the following options to install Kyma CLI from the latest release.

### Homebrew (macOS)

To install Kyma CLI using Homebrew, run:

```bash
brew install kyma-cli
```

### macOS

To install Kyma CLI on macOS, run:

```bash
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4)/kyma_Darwin_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && sudo mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz
```

### Linux

To install Kyma CLI on Linux, run:

```bash
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4)/kyma_Linux_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && sudo mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz
```

### Windows

To install Kyma CLI on Windows, download and unzip the [artifact](https://github.com/kyma-project/cli/releases). Remember to adjust your **PATH** environment variable.

```PowerShell
${KYMA_VERSION}=1.2.0
Invoke-WebRequest -OutFile kyma.zip https://github.com/kyma-project/cli/releases/download/${KYMA_VERSION}/kyma_Windows_x86_64.zip

Expand-Archive -Path kyma.zip -DestinationPath .\kyma-cli

cd kyma-cli
```

### Chocolatey (Windows)

To install Kyma CLI on Windows using [Chocolatey](https://www.chocolatey.org), run:

```PowerShell
choco install kyma-cli
```

### Other

To install a different release, change the path to point to the desired version and architecture:
```bash
curl -Lo kyma.tar.gz https://github.com/kyma-project/cli/releases/download/${KYMA_VERSION}/kyma_${ARCH}.tar.gz
```

## Usage

### Syntax

Use the following syntax to run the commands from your terminal:

```bash
kyma {COMMAND} {FLAGS}
```

where:

- **{COMMAND}** specifies the operation you want to perform.
- **{FLAGS}** specifies optional flags.

Example:

```bash
kyma install --source=latest
```

### Commands

Kyma CLI comes with a set of commands, each of which has its own specific set of flags. 

>**NOTE:** For the full list of commands and flags, see [this](https://github.com/kyma-project/cli/tree/master/docs/gen-docs) document. 

|     Command        | Child commands   |  Description  | Example |
|--------------------|----------------|---------------|---------|
| [`completion`](/docs/gen-docs/kyma_completion.md)| None| Generates and displays the bash or zsh completion script. | `kyma completion`|
| [`console`](/docs/gen-docs/kyma_console.md)| None| Launches Kyma Console in a browser window. | `kyma console` |
| [`install`](/docs/gen-docs/kyma_install.md)| None| Installs Kyma on a cluster based on the current or specified release. | `kyma install`|
| [`provision`](/docs/gen-docs/kyma_provision.md)| [`minikube`](/docs/gen-docs/kyma_provision_minikube.md)<br> [`gardener`](/docs/gen-docs/kyma_provision_gardener.md) <br> [`gcp`](/docs/gen-docs/kyma_provision_gcp.md) <br> [`azure`](/docs/gen-docs/kyma_provision_azure.md)| Provisions a new cluster on a platform of your choice. Currently, this command supports cluster provisioning on GCP, Azure, Gardener, and Minikube. | `kyma provision minikube`|
| [`test`](/docs/gen-docs/kyma_test.md)|[`definitions`](/docs/gen-docs/kyma_test_definitions.md)<br> [`delete`](/docs/gen-docs/kyma_test_delete.md) <br> [`list`](/docs/gen-docs/kyma_test_list.md) <br> [`run`](/docs/gen-docs/kyma_test_run.md) <br> [`status`](/docs/gen-docs/kyma_test_status.md)<br> [`logs`](/docs/gen-docs/kyma_test_logs.md) <br> | Runs and manages tests on a provisioned Kyma cluster. Using child commands, you can run tests, view test definitions, list and delete test suites, display test status, and fetch the logs of the tests.| `kyma test run` |
| [`version`](/docs/gen-docs/kyma_version.md)|None| Shows the cluster version and the Kyma CLI version.| `kyma version` |

### Usage examples

Further usage examples include:

- Install Kyma with Minikube on Mac:

    ```bash
    kyma provision minikube
    kyma install
    ```

- Install Kyma with Minikube on Windows:

    ```bash
    kyma provision minikube
    kyma install
    ```

- Install Kyma with Minikube on Windows using HyperV:

    ```bash
    kyma provision minikube --vm-driver hyperv --hypervVirtualSwitch {YOUR_SWITCH_NAME}
    kyma install

    ```

## Development

### Kyma CLI as a kubectl plugin

>**NOTE**: To use Kyma CLI as a kubectl plugin, use Kubernetes version 1.12.0 or higher.

A plugin is a standalone executable file with a name prefixed with `kubectl-` .To use the plugin, perform the following steps:

1. Rename the `kyma` binary to `kubectl-kyma` and place it anywhere in your **{PATH}**:

```bash
sudo mv ./kyma /usr/local/bin/kubectl-kyma
```

2. Run `kubectl plugin list` command to see your plugin on the list of all available plugins.

3. Invoke your plugin as a kubectl command:

```bash
$ kubectl kyma version
Kyma CLI version: v0.6.1
Kyma cluster version: 1.0.0
```

For more information on extending kubectl with plugins, read [Kubernetes documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).

### Kyma CLI stable binaries

Kyma CLI is used in continuous integration jobs that install or test Kyma or provision clusters. To effectively support this, we publish the stable binaries created from the `stable` tag which corresponds to the latest stable version of Kyma CLI.

To download the binaries, run:

```bash
curl -Lo kyma https://storage.googleapis.com/kyma-cli-stable/kyma-darwin # kyma-linux or kyma.exe
chmod +x kyma
```

### Kyma CLI Homebrew formula

If the Homebrew formula of the CLI does not get updated by the Homebrew team within three days of the release, update the formula of the CLI manually to the most recent version by following this [guide](https://github.com/Homebrew/brew/blob/master/docs/How-To-Open-a-Homebrew-Pull-Request.md). For a sample Homebrew Kyma CLI formula version bump, see [this](https://github.com/Homebrew/homebrew-core/pull/52375) PR.

### Kyma CLI Chocolatey package

The Kyma CLI Chocolatey package does not need to be bumped when there is a new release, as it has a script that will automatically check for new releases and update the package to the latest release.

Nevertheless, the package still needs some maintenance to keep its dedicated [site](https://chocolatey.org/packages/kyma-cli) at chocolatey.org up to date (e.g. update description, details, screenshots, etc...).

In order to maintain the [site](https://chocolatey.org/packages/kyma-cli), send a pull request to Chocolatey's [GitHub repository](https://github.com/dgalbraith/chocolatey-packages/tree/master/automatic/kyma-cli).
