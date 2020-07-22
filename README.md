# Kyma CLI

## Overview

Kyma CLI is a command line tool which supports [Kyma](https://github.com/kyma-project/kyma) developers. It provides a set of commands you can use to provision clusters as well as install, manage, and test Kyma.

>**TIP:** This document briefly describes the concept of Kyma CLI. Read the [complete documentation](https://github.com/kyma-project/cli/blob/master/docs) to learn more.

## Quick installation

>**NOTE:** Before you start with the installation, make sure you have [Minikube](https://github.com/kubernetes/minikube) installed.

Use the following options to install Kyma CLI from the latest release.

To install Kyma CLI on MacOS using Homebrew, run:

```bash
brew install kyma-cli
```

To install Kyma CLI on Windows using [Chocolatey](https://www.chocolatey.org), run:

```PowerShell
choco install kyma-cli
```
Read more on [installation options](https://github.com/kyma-project/cli/blob/master/docs/03-01-installation.md).

## Use Kyma CLI

Once you have already installed the CLI, you can use its set of handy commands and flags to set up your cluster locally or on one of the cloud providers, install Kyma, manage and test it.
In a nutshell, this is the syntax you can use to run the commands from your terminal:

```bash
kyma {COMMAND} {FLAGS}
```

where:

- **{COMMAND}** specifies the operation you want to perform.
- **{FLAGS}** specifies optional flags.

Example:

```bash
kyma install -s latest
```

>**TIP:** 

>Read more about the available [commands and flags](https://github.com/kyma-project/cli/blob/master/docs/03-02-use-kyma-cli.md).
>You can also try out Kyma CLI using [these examples](https://github.com/kyma-project/cli/blob/master/docs/03-03-examples.md).

## Development

### Kyma CLI stable binaries

Kyma CLI is used in CI jobs that install or test Kyma or provision clusters. To effectively support this process, we publish the stable binaries created from the `stable` tag which corresponds to the latest stable version of Kyma CLI.

To download the binaries, run:

```bash
curl -Lo kyma https://storage.googleapis.com/kyma-cli-stable/kyma-darwin # kyma-linux or kyma.exe
chmod +x kyma
```

### Kyma CLI as kubectl plugin

To learn how to use Kyma CLI as a kubectl plugin, follow [this tutorial](https://github.com/kyma-project/cli/blob/master/docs/09-01-kubectl-plugin-tutorial.md).