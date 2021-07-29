---
title: kyma alpha deploy
---

Deploys Kyma on a running Kubernetes cluster.

## Synopsis

Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.

Usage Examples:
  Deploy Kyma using your own domain name
    You must provide the certificate and key as files.
    If you don't have a certificate yet, you can create a self-signed certificate and key:
		openssl req -x509 -newkey rsa:4096 -keyout key.pem -out crt.pem -days 365
    Then, pass the certificate files to the deploy command:
		kyma alpha deploy --domain {DOMAIN} --tls-cert crt.pem --tls-key key.pem

  Deploy Kyma from specific source:
    - Deploy from a specific version, such as 1.19.1:
		kyma alpha deploy --source=1.19.1
    - Build Kyma from local sources and deploy on remote cluster:
		kyma alpha deploy --source=local

  Deploy Kyma with only specific components:
    You need to pass a path to a YAML file containing desired components. An example YAML file would contain:
		prerequisites:
		- name: "cluster-essentials"
		- name: "istio"
		  namespace: "istio-system"
		components:
		- name: "testing"
		- name: "xip-patch"
		- name: "istio-kyma-patch"
		- name: "dex"
    Then run:
		kyma alpha deploy --components {COMPONENTS_FILE_PATH}

  Change Kyma settings:
    To change your Kyma configuration, use the alpha deploy command and deploy the same Kyma version that you're currently using,
    just with different settings.
    - Using a settings-file:
		kyma alpha deploy --values-file {VALUES_FILE_PATH}
    - Using specific values instead of file:
		kyma deploy --value ory.hydra.deployment.resources.limits.cpu=153m \
		--value ory.hydra.deployment.resources.requests.cpu=53m

Debugging:
  The alpha commands support troubleshooting in several ways, for example:
  - To get a detailed view of the installation process, use the --verbose flag.
  - To tweak the values on a component level, use alpha deploy --components:
    Pass a components list that includes only the components you want to test
    and try out the settings that work for your installation.
  - To understand which component failed during deployment, deactivate the default atomic deployment:
		--atomic=false
    With atomic deployment active, any component that hasn't been installed successfully is rolled back,
    which may make it hard to find out what went wrong. By disabling the flag, the failed components are not rolled back.
	

```bash
kyma alpha deploy [flags]
```

## Flags

```bash
  -a, --atomic                       Set --atomic=true to use atomic deployment, which rolls back any component that could not be installed successfully.
      --component strings            Provide one or more components to deploy (e.g. --component componentName@namespace)
  -c, --components-file string       Path to the components file (default "$HOME/.kyma/sources/installation/resources/components.yaml" or ".kyma-sources/installation/resources/components.yaml")
      --concurrency int              Number of parallel processes (default 4)
  -d, --domain string                Custom domain used for installation
  -p, --profile string               Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: "evaluation", "production".
  -r, --reuse-values                 Set --reuse-values=true to reuse the helm values during component installation
  -s, --source string                Installation source:
                                     	- Deploy a specific release, for example: "kyma alpha deploy --source=1.17.1"
                                     	- Deploy a specific branch of the Kyma repository on kyma-project.org: "kyma alpha deploy --source=<my-branch-name>"
                                     	- Deploy a commit, for example: "kyma alpha deploy --source=34edf09a"
                                     	- Deploy a pull request, for example "kyma alpha deploy --source=PR-9486"
                                     	- Deploy the local sources: "kyma alpha deploy --source=local" (default "main")
      --timeout duration             Maximum time for the deployment (default 20m0s)
      --timeout-component duration   Maximum time to deploy the component (default 6m0s)
      --tls-crt string               TLS certificate file for the domain used for installation
      --tls-key string               TLS key file for the domain used for installation
      --value strings                Set configuration values. Can specify one or more values, also as a comma-separated list (e.g. --value component.a='1' --value component.b='2' or --value component.a='1',component.b='2').
  -f, --values-file strings          Path(s) to one or more JSON or YAML files with configuration values
  -w, --workspace string             Path to download Kyma sources (default "$HOME/.kyma/sources" or ".kyma-sources")
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

* [kyma alpha](#kyma-alpha-kyma-alpha)	 - Executes the commands in the alpha testing stage.

