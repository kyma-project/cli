---
title: kyma install
---

Installs Kyma on a running Kubernetes cluster.

## Synopsis

Use this command to install Kyma on a running Kubernetes cluster.
		
### Description

Before you use the command, make sure your setup meets the following prerequisites:

* Kyma is not installed.
* A Kubernetes cluster is available with your kubeconfig file already pointing to it.

Here are the installation steps:

The standard installation uses the minimal configuration. 
Depending on your installation type, the ways to deploy and configure the Kyma Installer are different:

If you install Kyma locally **from a release**, the system does the following:

   1. Fetch the latest or specified release, along with configuration.
   2. Deploy the Kyma Installer on the cluster.
   3. Apply the downloaded or defined configuration.
   
If you install Kyma locally **from sources**, the system does the following:

   1. Fetch the configuration yaml files from the local sources.
   2. Build the Kyma Installer image.
   3. Deploy the Kyma Installer and apply the fetched configuration.
   
Both installation types continue with the following steps:
   
   4. If overrides have been defined, apply them.
   5. Set the admin password.
   6. Patche the Minikube IP.
   

```bash
kyma install [flags]
```

## Flags

```bash
  -c, --components string      Path to a YAML file with a component list to override.
      --custom-image string    Full image name including the registry and the tag. Required for installation from local sources to a remote cluster.
  -d, --domain string          Domain used for installation. (default "kyma.local")
      --fallback-level int     If "source=main", defines the number of commits from main branch taken into account if artifacts for newer commits do not exist yet (default 5)
  -n, --no-wait                Determines if the command should wait for Kyma installation to complete.
  -o, --override stringArray   Path to a YAML file with parameters to override.
  -p, --password string        Predefined cluster password.
      --profile string         Kyma installation profile (evaluation|production). If not specified, Kyma is installed with the default chart values.
  -s, --source string          Installation source.
                               	- To use a specific release, write "kyma install --source=1.15.1".
                               	- To use the main branch, write "kyma install --source=main".
                               	- To use a commit, write "kyma install --source=34edf09a".
                               	- To use a pull request, write "kyma install --source=PR-9486".
                               	- To use the local sources, write "kyma install --source=local".
                               	- To use a custom installer image, write "kyma install --source=user/my-kyma-installer:v1.4.0".
      --src-path string        Absolute path to local sources.
      --timeout duration       Timeout after which CLI stops watching the installation progress. (default 1h0m0s)
      --tls-cert string        TLS certificate for the domain used for installation. The certificate must be a base64-encoded value.
      --tls-key string         TLS key for the domain used for installation. The key must be a base64-encoded value.
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Command help
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.

