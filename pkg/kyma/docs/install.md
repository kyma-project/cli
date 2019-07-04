# install

## Description

Installs Kyma on a running Kubernetes cluster.

## Usage

```
kyma install [OPTIONS]
```

## Options

| Name     | Short Name | Default value| Description|
| ----------|---------|-----|------|
| --release | -r ||Specifies the Kyma release or git revision to be installed. Go to the [GitHub releases page](https://github.com/kyma-project/kyma/releases) to find out more about each of the available releases, or use the revision of your choice. For example, `kyma install --release master`.|
| --local | -r |false|Indicates installtion from sources. To make it work, make sure you have **{GO_PATH}** set and the location of your sources complies with the Go file and naming convention.| 
| --config | -r ||Specifies the URL or path to the Installer configuration YAML file.| 
| --override | -r ||Specifies the path to YAML file with parameters to override. Multiple entries of this flag are allowed.| 
| --domain | -r |kyma.local|Specifies the domain used for installation.| 
| --password | -r ||Specifies the predefined cluster password.| 
| --noWait | -r |false|Determines if the installation should wait for the Installer configuration to complete.| 
| --src-path | ||Specifies the absolute path to local sources.| 
| --installer-version | ||Specifies the version of the Kyma Installer Docker image used for the local installation.| 
| --installer-dir | ||Specifies the directory of the Kyma Installer Docker image used for the local installation.| 
| --timeout |  |0|Specifies the timeout after which CLI stops watching the installation progress.| 

## Detailed description 

### Prerequisites
- Kyma is not installed.
- Kubernetes cluster is available with your KUBECONFIG already pointing to it.
- Helm binary is available (optional).

### Installation flow 

During the installation, the system performs the following steps:
1. Fetches the tiller.yaml file from /installation/resources directory and deploys it to the cluster.
2. Deploys and configures the Kyma Installer. This is a standard installation using the latest minimal configuration. 
You can override the settings using the --override or --config flag.
  2a) If you choose to install Kyma from release, the system:
	 - Fetches the latest or specified release along with configuration
	 - Deploys the Kyma Installer on the cluster
	 - Applies downloaded or defined configuration
	 - Applies overrides if applicable
	 - Sets admin password
	 - Patches the minikube IP
  2b) If you choose to install Kyma from local sources, the system:
	 - Fetches local resources YAML files
	 - Builds the Kyma Installer image
	 - Deploys the Kyma Installer and applies the fetched configuration
	 - Applies overrides if applicable
	 - Sets admin password
	 - Patches minikube IP
3. Configures Helm (optional). 
   If installed, Helm is automatically configured using certificates from tiller.
4. Runs Kyma installation until the status `installed` confirms the success.

## Examples

The following examples include the most common cases of using the install command. 
1. Install Kyma from the current release:
   ```bash
   kyma install
   ```

2. Install Kyma from sources:
   ```bash
   kyma install --local
   ```
3. Install Kyma using your own configuration. [Here](https://github.com/kyma-project/kyma/releases/download/1.2.2/kyma-installer-local.yaml) you can find an example of the Installer configuration file you can base your own configuration on.
   ```bash
   kyma install --config {YAML_FILE_PATH}
   ```
4. Install Kyma and override specific parameters. For more detailson overrides, see [this](https://kyma-project.io/docs/root/kyma#configuration-helm-overrides-for-kyma-installation) document.
   ```bash
   kyma install --override {YAML_FILE_PATH}
   ```