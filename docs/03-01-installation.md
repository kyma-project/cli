---
title: Install Kyma CLI
type: Details
---

You can easily install Kyma CLI on macOS, Linux, or Windows. Follow the instructions described in the sections.

## Homebrew

To install Kyma CLI using Homebrew, run:

```bash
brew install kyma-cli
```

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

## Other

To install a different release version, change the path to point to the desired version and architecture:

```bash
curl -Lo kyma.tar.gz https://github.com/kyma-project/cli/releases/download/${KYMA_VERSION}/kyma_${ARCH}.tar.gz
```