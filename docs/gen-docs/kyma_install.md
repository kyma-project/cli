## kyma install

Installs Kyma on a running Kubernetes cluster.

### Synopsis

Use this command to install Kyma on a running Kubernetes cluster.

### Detailed description

Before you use the command, make sure your setup meets the following prerequisites:

* Kyma is not installed.
* Kubernetes cluster is available with your kubeconfig file already pointing to it.
* Helm binary is available (optional).

Here are the installation steps:

The standard installation uses the minimal configuration. The system performs the following steps:
1. Fetches the `tiller.yaml` file from the `/installation/resources` directory and deploys it to the cluster.
2. Deploys and configures the Kyma Installer. At this point, steps differ depending on the installation type.
    <div tabs name="installation">
    <details>
    <summary>
    From release
    </summary>

    When you install Kyma locally from release, the system:
    1. Fetches the latest or specified release along with configuration.
    2. Deploys the Kyma Installer on the cluster.
    3. Applies downloaded or defined configuration.
    4. Applies overrides, if applicable.
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
4. Runs Kyma installation until the **installed** status confirms the successful installation.
	> **NOTE**: You can override the standard installation settings using the `--override` or `--config` flag.

### Usage


```
kyma install [flags]
```

### Options

```
  -d, --domain string       Specifies the domain used for installation. (default "kyma.local")
  -h, --help                Displays help for the command.
  -n, --noWait              Determines if the command should wait for the Kyma installation to complete.
  -o, --override []string   Specifies the path to a yaml file with parameters to override. (default [])
  -p, --password string     Specifies the predefined cluster password.
  -s, --source string       Specifies the installation source. To use the specific release, write kyma install --source=1.3.0. To use the latest master, write kyma install --source=latest. To use the local sources, write kyma install --source=local. To use the remote image, write kyma install --source=user/my-kyma-installer:v1.4.0
      --src-path string     Specifies the absolute path to local sources.
      --timeout duration    Time-out after which CLI stops watching the installation progress (default 30m0s)
      --tlsCert string      Specifies the TLS certificate for the domain used for installation.
      --tlsKey string       Specifies the TLS key for the domain used for installation.
```

### Options inherited from parent commands

```
      --kubeconfig string   Specifies the path to the kubeconfig file. (default "/Users/user/.kube/config")
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

### SEE ALSO

* [kyma](kyma.md)	 - Controls a Kyma cluster.

