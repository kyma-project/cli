# Kyma-CLI

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

For the latest relase and installation instructions, see the [release page](https://github.com/kyma-incubator/kyma-cli/releases/tag/v0.3.0)

## Kyma-CLI as a Kubectl plugin

To follow this section a kubectl version of 1.12.0 or later is required.

A plugin is nothing more than a standalone executable file, whose name begins with kubectl- . To install a plugin, simply move this executable file to anywhere on your PATH.

Rename a `kyma` binary to `kubectl-kyma` and place it anywhere in your PATH:

```bash
sudo mv ./kyma /usr/local/bin/kubectl-kyma
```

Run `kubectl plugin list` command and you will see your plugin in the list of available plugins.

You may now invoke your plugin as a kubectl command:

```bash
$ kubectl kyma status
Kyma is running!
```

To know more about extending kubectl with plugins read [kubernetes documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).

## Roadmap

- Portability
  - adding windows support (only hosts manipulation missing)
  - validate linux support
- Cloud Provider support
  - Google Kubernetes Engine
- Kyma installation
  - open the Kyma console in the default browser at the end of Kyma installation
  - provide own command "dashboard" to open the conole in default browser
  - support all configuration options of the kyma-installer
  - install optional kyma module (ark, logging)
  - uninstall optional kyma module
  - update kyma to newer release
- Release management
  - SUpport for old releases
  - use latest release automically
  - list available releases
- Application Connectivity
  - create remote environment and fetch connection token
  - manage APIs registered by an application
- Testing/Validation/Debugging
  - connect mock application to kyma
  - 'Check' a kyma installation for potential problems
  - Query logs of a pod/namespace
- CTL installation
  - homebrew support
  - support for edge releases
- Engineering
  - use log framework with log levels and have verbose modes
  - have tests
  - use kubernetes go-client instead of kubectl command execution
  - improve and review help texts
- Service Catalog
  - support for ServiceBindingUsage
  
