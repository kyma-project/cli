# install

## Description

Use this command to install Kyma on a running Kubernetes cluster.

## Usage

```
kyma install [OPTIONS]
```

## Options

| Name                    | Short Name | Default value| Description|
| :----------------------:|:---------:|:-----:|------|
| **&#x2011;&#x2011;release**  | -r ||Specifies the Kyma release or Git revision to be installed. Go to the [GitHub releases page](https://github.com/kyma-project/kyma/releases) to find out more about each of the available releases, or use the revision of your choice. For example, write `kyma install --release master`.|
| **&#x2011;&#x2011;local**    | -l |`false`|Indicates local installation using Kyma sources. If the location of your cloned  `kyma-cli` repository follows the Go code conventions, the CLI finds it automatically. If not, you must configure the path explicitly using `--src-path`.| 
| **&#x2011;&#x2011;config**   | -c ||Specifies the URL or path to the Installer configuration yaml file.| 
| **&#x2011;&#x2011;override** | -o ||Specifies the path to a yaml file with parameters to override.| 
| **&#x2011;&#x2011;domain**   | -d |`kyma.local`|Specifies the domain used for installation.| 
| **&#x2011;&#x2011;password** | -p ||Specifies the predefined cluster password.| 
| **&#x2011;&#x2011;noWait**   | -n |`false`|Determines if the command should wait for the should wait for the Kyma installation to complete.| 
| **&#x2011;&#x2011;src&#x2011;path**  | ||Specifies the absolute path to local sources.| 
| **&#x2011;&#x2011;installer&#x2011;version** | ||Specifies the version of the Kyma Installer Docker image used for the local installation.| 
| **&#x2011;&#x2011;installer&#x2011;dir**     | ||Specifies the directory of the Kyma Installer Docker image used for the local installation.| 
| **&#x2011;&#x2011;timeout**           |  |`30m`|Specifies the time-out after which the CLI stops watching the installation progress. The time-out value is a [duration string](https://golang.org/pkg/time/#ParseDuration) meaning a decimal number followed by a unit suffix.| 

## Details

Learn more about the actions triggered by the command.

### Prerequisites

Before you use the command, make sure your setup meets the following prerequisites:

* Kyma is not installed.
* Kubernetes cluster is available with your KUBECONFIG already pointing to it.
* Helm binary is available (optional).

### Installation flow 

The standard installation uses the minimal configuration. The system performs the following steps:
1. Fetches the `tiller.yaml` file from the `/installation/resources` directory and deploys it to the cluster.
2. Deploys and configures the Kyma Installer. At this point, the steps differ depending on the installation type.
    <div tabs name="installation">
    <details>
    <summary>
    From release
    </summary>

    When you install Kyma locally from release, the system:
    1. Fetches the latest or specified release along with configuration.
    2. Deploys the Kyma Installer on the cluster.
    3. Applies downloaded or defined configuration.
    4. Applies overrides if applicable.
    5. Sets the admin password.
    6. Patches the Minikube IP.
    </details>
    <details>
    <summary>
    From sources
    </summary>
    
    When you install Kyma locally from sources, the system:
    1. Fetches the configuration yaml files from the local sources.
    2. Builds the Kyma Installer image.
    3. Deploys the Kyma Installer and applies the fetched configuration.
    4. Applies overrides, if applicable.
    5. Sets the admin password.
    6. Patches the Minikube IP.
    </details>
    </div>
3. Configures Helm. If installed, Helm is automatically configured using certificates from Tiller. This step is optional.
4. Runs Kyma installation until the `installed` status confirms the successful installation.
> **NOTE**: You can override the standard installation settings using the `--override` or `--config` flag.

## Examples

The following examples include the most common uses of the `install` command. 
* Install Kyma from the current release:
   ```bash
   kyma install
   ```
* Install Kyma from local sources:

   >**NOTE**: The location of your cloned `kyma-cli` repository must comply with Go code naming conventions. 

   ```bash
   kyma install --local
   ```
* Install Kyma from local sources using the absolute **{SRC_PATH}**:
   ```bash
   kyma install --src-path {SRC_PATH}
   ```
* Install Kyma using your own configuration:

   ```bash
   kyma install --config {yaml_FILE_PATH}
   ```
   [Here](https://github.com/kyma-project/kyma/releases/download/1.2.2/kyma-installer-local.yaml) you can find an example of the Kyma Operator configuration yaml file. Use it to create your own configuration.

* Install Kyma and override specific parameters:

   ```bash
   kyma install --override {yaml_FILE_PATH}
   ```
   For details on overrides, see [this](https://kyma-project.io/docs/root/kyma#configuration-helm-overrides-for-kyma-installation) document. 