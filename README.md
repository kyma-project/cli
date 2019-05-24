# Kyma CLI

## Overview

Kyma CLI is a command line tool which supports [Kyma](https://github.com/kyma-project/kyma) developers. It features a set of commands to facilitate installing and managing Kyma.

## Prerequisites

Kyma CLI requires the following software:
- [kubectl](https://github.com/kubernetes/kubectl) 
- [minikube](https://github.com/kubernetes/minikube) 

## Installation

For the latest release and installation instructions, see the [release page](https://github.com/kyma-project/cli/releases)

## Usage

### Available commands

Kyma CLI comes with a set of commands:

- `version`: Shows the Kyma cluster version and the Kyma CLI version
- `provision minikube`: Initializes Minikube with a new cluster (replaces the `minikube.sh` script) 
- `install`: Installs Kyma to a cluster based on a release (replaces the `ìnstaller.sh` and `is-installed.sh` script)
- `uninstall`: Uninstalls all Kyma related resources from a cluster
- `completion`: Outputs shell completion code for bash
- `test`: Triggers and reports the tests for every Kyma module
- `help`: Displays and explains usage of a given command, for example, `kyma help`, `kyma help status`

### Use Kyma CLI

Install Kyma with Minikube on Mac:

```bash
kyma provision minikube
kyma install
```

Install Kyma with Minikube on Windows:

```bash
kyma provision minikube
# follow instructions to add hosts
kyma install
```

Install Kyma with Minikube on Windows using HyperV:

```bash
kyma provision minikube --vm-driver hyperv --hypervVirtualSwitch {YOUR_SWITCH_NAME}
# follow instructions to add hosts
kyma install
```

Run tests on Kyma:
```bash
kyma test
```

## Development

### Kyma CLI as a Kubectl plugin

To follow this section a kubectl version of 1.12.0 or later is required.

A plugin is nothing more than a standalone executable file, whose name begins with kubectl- . To install a plugin, simply move this executable file to anywhere on your PATH.

Rename a `kyma` binary to `kubectl-kyma` and place it anywhere in your PATH:

```bash
sudo mv ./kyma /usr/local/bin/kubectl-kyma
```

Run `kubectl plugin list` command and you will see your plugin in the list of available plugins.

You may now invoke your plugin as a kubectl command:

```bash
$ kubectl kyma version
Kyma CLI version: v0.6.1
Kyma cluster version: 1.0.0
```

To know more about extending kubectl with plugins read [Kubernetes documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).
