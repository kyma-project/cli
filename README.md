# Kyma CLI

## Overview

Kyma CLI is a command line tool which supports [Kyma](https://github.com/kyma-project/kyma) developers. It provides a set of commands you can use to provision clusters as well as install, manage, and test Kyma.

>**TIP:** This document briefly describes the concept of Kyma CLI. Read the [complete documentation](https://kyma-project.io/docs/cli/) to learn more.

## Installation

To install the latest release of Kyma CLI on MacOS using Homebrew, run:

```bash
brew install kyma-cli
```

To install the latest release of Kyma CLI on Windows using [Chocolatey](https://www.chocolatey.org), run:

```PowerShell
choco install kyma-cli
```

Read more about [installation options](https://github.com/kyma-project/cli/blob/master/docs/04-01-installation.md).

## Use Kyma CLI

Once you have installed the CLI, you can use its set of commands and flags to provision a cluster and start working with Kyma.

In a nutshell, this is the syntax you can use to run the commands from your terminal:

```bash
kyma {COMMAND} {FLAGS}
```

where:

- **{COMMAND}** specifies the operation you want to perform.
- **{FLAGS}** specifies optional flags.

Example:

```bash
kyma install -s master
```

>**TIP:** Read more about the available [commands and flags](https://github.com/kyma-project/cli/blob/master/docs/03-01-use-kyma-cli.md).You can also try out Kyma CLI using [these examples](https://github.com/kyma-project/cli/blob/master/docs/03-02-examples.md).

### Kyma CLI as kubectl plugin

To learn how to use Kyma CLI as a kubectl plugin, follow [this tutorial](https://github.com/kyma-project/cli/blob/master/docs/08-01-kubectl-plugin-tutorial.md).

## Development

### Build from sources

You can build the Kyma CLI from the sources:

1. To clone the Kyma CLI repository, run:

```bash
git clone git@github.com:kyma-project/cli.git
```

2. Enter the root folder of the cloned repository:

```bash
cd cli
```

3. Run `make build` for your target OS:
    - Mac OSX: `make build-darwin`
    - Windows: `make build-windows`
    - Linux: `make build-linux`

The binary is saved to the `bin` folder in the Kyma CLI repository.

### Kyma CLI stable binaries

Kyma CLI is used in CI jobs that install or test Kyma or provision clusters. To effectively support this process, we publish the stable binaries created from the `stable` tag which corresponds to the latest stable version of Kyma CLI.

To download the binaries, run:

```bash
curl -Lo kyma https://storage.googleapis.com/kyma-cli-stable/kyma-darwin # kyma-linux or kyma.exe
chmod +x kyma
```