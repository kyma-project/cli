# Kymactl

## Overview

A command line tool to support developers of and with Kyma

## Available Commands

- `version`: Shows the kyma cluster version and the Kyma CLI version. The Kyma CLI version is set at compile time passing it to the go linker as a flag:

    ```bash
    go build -o ./bin/kyma -ldflags '-X github.com/kyma-incubator/kymactl/pkg/kyma/cmd.Version=1.5.0' ./cmd/kyma
    ```
- `install cluster minikube`: Initializes minikube with a new cluster (replaces the `minikube.sh` script) 
- `install kyma`: Installs kyma to a cluster based on a release (replaces the `ìnstaller.sh` and `is-installed.sh` script)
- `uninstall kyma`: Uninstalls all kyma related resources from a cluster
- `completion`: Output shell completion code for bash.
- `help`: Displays usage for the given command (e.g. `kyma help`, `kyma help status`, etc...)

## Usage

Installation of kyma with minikube on Mac:

```bash
kyma install cluster minikube
kyma install kyma
```

Installation of kyma with minikube on Windows:

```bash
kyma install cluster minikube --vm-driver hyperv --hypervVirtualSwitch {YOUR_SWITCH_NAME}
# follow instructions to add hosts

kyma install kyma
```

Run tests on Kyma installation:
```bash
kyma test
```

## Kymactl as a Kubectl plugin

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

- Usability
  - Better Command Structure?:
    - kyma install
    - kyma uninstall
    - kyma add monitoring
    - kyma cluster minikube
    - kyma cluster gke
    - kyma version
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
  - use latest release automically
  - list available releases
- Application Connectivity
  - create remote environment and fetch connection token
  - manage APIs registered by an application
- Testing/Validation
  - execute acceptance tests for kyma
  - connect mock application to kyma
  - 'Check' a kyma installation for potential problems
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
  
