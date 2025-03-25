# Kyma CLI

> [!WARNING]
> This documentation applies to Kyma CLI version `v3`, which is still in the prototyping stage. All commands are still considered alpha. Use it at your own risk.
> Read more about the decision [here](https://github.com/kyma-project/community/issues/872).

## What is Kyma CLI?

The Kyma CLI is a command-line interface tool designed to simplify the use of the Kyma ecosystem. It allows you to manage Kyma modules and applications, deploy simple applications, and more. With Kyma CLI, you can perform complex tasks with simple commands, accelerating development cycles.

In addition, you can build, push, and deploy an application to a Kyma cluster with a single command. It automatically detects the Dockerfile in the current directory, builds and pushes the image to the in-cluster registry, and applies the necessary Kubernetes resources.

Kyma CLI also provides a set of commands to manage Kyma modules efficiently. You can manage, deploy, and configure modules seamlessly. With Kyma CLI module commands, you can list available and installed modules, add or delete them, and configure their settings. Modules can be deployed with the default or custom configuration.

## Features

The Kyma CLI provides the following features:

- Simplified module management.
- Automated deployments.
- Built-in extensibility.
- Integrated service management.
- Commands providing useful automation.

## How to Install

> [!WARNING]
> `v3` is still in the prototyping stage. All commands are still considered alpha. Use it at your own risk.

### Stable Release

To get the latest stable Kyma CLI for MacOS or Linux, run the following script from the command line:

```sh
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_stable.sh" | sh -
kyma
```

### Latest Build

Download the latest (stable or unstable pre-release) v3 build from the [releases](https://github.com/kyma-project/cli/releases) assets.

To get the latest Kyma CLI for MacOS or Linux, run the following script from the command line:

```sh
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_latest.sh" | sh -
kyma
```

### Nightly Build

Download the latest build from the main branch from [0.0.0-dev](https://github.com/kyma-project/cli/releases/tag/0.0.0-dev) release assets.

To get Kyma CLI for MacOS or Linux, run the following script from the command line:

```sh
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_nightly.sh" | sh -
kyma
```

## Related Information

- [Kyma CLI tutorials](tutorials/README.md)
