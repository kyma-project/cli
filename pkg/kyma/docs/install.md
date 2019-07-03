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
| --release | -r ||Kyma release to use|
| --local | -r |false|Install from sources| 
| --config | -r ||URL or path to the Installer configuration YAML file| 
| --override | -r ||Path to YAML file with parameters to override. Multiple entries of this flag are allowed| 
| --domain | -r |kyma.local|Domain used for installation| 
| --password | -r ||Predefined cluster password| 
| --noWait | -r |false|Do not wait for the Installer configuration to complete| 
| --src-path | ||Path to local sources| 
| --installer-version | ||Version of the Kyma Installer Docker image used for local installation| 
| --installer-dir | ||The directory of the Kyma Installer Docker image used for local installation| 
| --timeout |  |0|Timeout after which CLI stops watching the installation progress| 

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
3. Configures Helm (optional). If installed, Helm is automatically configured using certificates from tiller.
4. Runs Kyma installation until the status "Installed" confirms the success.

## Examples

The following examples include the most common cases of using the install command. 
1. Install Kyma from the current release:
   ```bash
   install kyma
   ```

2. Install Kyma from sources:
   ```bash
   install kyma --local
   ```

3. Install Kyma using your own configuration:
   ```bash
   install kyma --config {YAML_FILE_PATH}
   ```
4. Install Kyma and override parameters listed in a YAML file:
   ```bash
   install kyma --override {YAML_FILE_PATH}
   ```