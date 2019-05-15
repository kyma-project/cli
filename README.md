# Kyma CLI

## Overview

A command line tool to support developers of and with [Kyma](https://github.com/kyma-project/kyma)

## Available Commands

- `version`: Shows the Kyma cluster version and the Kyma CLI version
- `provision minikube`: Initializes minikube with a new cluster (replaces the `minikube.sh` script) 
- `install`: Installs Kyma to a cluster based on a release (replaces the `ìnstaller.sh` and `is-installed.sh` script)
- `uninstall`: Uninstalls all Kyma related resources from a cluster
- `completion`: Outputs shell completion code for bash
- `test`: Triggers and reports the tests for every Kyma module
- `help`: Displays usage for the given command (e.g. `kyma help`, `kyma help status`, etc...)

## Prerequisites

In order to run the Kyma CLI you need the following software installed:
- [kubectl](https://github.com/kubernetes/kubectl) 
- [minikube](https://github.com/kubernetes/minikube) 

## Usage

Installation of Kyma with minikube on Mac:

```bash
kyma provision minikube
kyma install
```

Installation of Kyma with minikube on Windows:

```bash
kyma provision minikube
# follow instructions to add hosts
kyma install
```

Installation of Kyma with minikube on Windows using HyperV:

```bash
kyma provision minikube --vm-driver hyperv --hypervVirtualSwitch {YOUR_SWITCH_NAME}
# follow instructions to add hosts
kyma install
```

Run tests on Kyma installation:
```bash
kyma test
```
## Installation

For the latest relase and installation instructions, see the [release page](https://github.com/kyma-project/cli/releases)

## Kyma CLI as a Kubectl plugin

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

To know more about extending kubectl with plugins read [kubernetes documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).

Testing Prow...
