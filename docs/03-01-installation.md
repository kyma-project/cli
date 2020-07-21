---
title: Install Kyma CLI
type: Details
---

You can easily install Kyma CLI on macOS, Linux, or Windows. Follow the instructions described in the sections.

## Prerequisites

To work, Kyma CLI requires the following software:

- [Minikube](https://github.com/kubernetes/minikube)

## Homebrew

To install Kyma CLI using Homebrew, run:

```bash
brew install kyma-cli
```

If the Homebrew formula of the CLI does not get updated by the Homebrew team within three days of the release, update the formula of the CLI manually to the most recent version by following this [guide](https://github.com/Homebrew/brew/blob/master/docs/How-To-Open-a-Homebrew-Pull-Request.md). For a sample Homebrew Kyma CLI formula version bump, see [this](https://github.com/Homebrew/homebrew-core/pull/52375) PR.

## macOS

To install Kyma CLI on macOS, run:

```bash
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4)/kyma_Darwin_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && sudo mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz
```

## Linux

To install Kyma CLI on Linux, run:

```bash
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4)/kyma_Linux_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && sudo mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz
```

## Windows

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
The Kyma CLI Chocolatey package does not need to be bumped when there is a new release, as it has a script that will automatically check for new releases and update the package to the latest release.

Nevertheless, the package still needs some maintenance to keep its dedicated [site](https://chocolatey.org/packages/kyma-cli) at chocolatey.org up to date (e.g. update description, details, screenshots, etc...).

In order to maintain the [site](https://chocolatey.org/packages/kyma-cli), send a pull request to Chocolatey's [GitHub repository](https://github.com/dgalbraith/chocolatey-packages/tree/master/automatic/kyma-cli).

## Other

To install a different release version, change the path to point to the desired version and architecture:

```bash
curl -Lo kyma.tar.gz https://github.com/kyma-project/cli/releases/download/${KYMA_VERSION}/kyma_${ARCH}.tar.gz
```