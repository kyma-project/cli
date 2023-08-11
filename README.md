# Kyma CLI

## Overview

Kyma CLI is a command line tool which supports [Kyma](https://github.com/kyma-project/kyma) developers. It provides a set of commands and flags you can use to:

- Provision a cluster locally or on cloud providers, such as GCP or Azure, or use Gardener to set up and easily manage your clusters.
- Deploy, manage, and undeploy Kyma.
- Manage your Functions.

>**TIP:** This document briefly describes the concept of Kyma CLI. Read the [Kyma documentation](https://kyma-project.io/#/01-overview/ui/README?id=kyma-cli) to learn more.

## Installation

To install the latest release of Kyma CLI on MacOS using Homebrew, run:

```bash
brew install kyma-cli
```

To install the latest release of Kyma CLI on Windows using [Chocolatey](https://www.chocolatey.org), run:

```PowerShell
choco install kyma-cli
```

Read more about [installation options](https://kyma-project.io/#/04-operation-guides/operations/01-install-kyma-CLI).

## Use Kyma CLI

Once you have installed the CLI, you can use its set of commands and flags to provision a cluster and start working with Kyma.

For the commands and flags to work, they must follow this syntax:

```bash
kyma {COMMAND} {FLAGS}
```

- **{COMMAND}** specifies the operation you want to perform, such as provisioning the cluster or deploying Kyma.
- **{FLAGS}** specifies optional flags you can use to enrich your command.

See the example:

```bash
kyma deploy -s main
```

>**TIP:** Read more about the available [commands and flags](https://github.com/kyma-project/cli/tree/main/docs/gen-docs).

## Development

### Build from Sources

Alternatively, you can also build the Kyma CLI from the sources:

1. To clone the Kyma CLI repository, run:

```bash
mkdir -p $GOPATH/src/github.com/kyma-project/
git clone git@github.com:kyma-project/cli.git $GOPATH/src/github.com/kyma-project/cli
```

2. Enter the root folder of the cloned repository:

```bash
cd $GOPATH/src/github.com/kyma-project/cli
```

3. Run `make build` for your target OS:
    - Mac OSX: `make build-darwin`
    - Windows: `make build-windows`
    - Linux: `make build-linux`

The binary is saved to the `bin` folder in the Kyma CLI repository.

### Kyma CLI binaries

Kyma CLI is used in CI jobs that install or test Kyma or provision clusters. To effectively support this process, 
we provide binaries using the latest reconciler version. The binaries are built every time a PR is merged into the main branch. We publish the binaries as unstable images.

To download the binaries using the latest reconciler version, run:

```bash
curl -Lo kyma https://storage.googleapis.com/kyma-cli-unstable/kyma-darwin # kyma-linux, kyma-linux-arm, kyma.exe, or kyma-arm.exe
chmod +x kyma
```
