---
title: kyma alpha deploy
---

Deploys Kyma on a running Kubernetes cluster.

## Synopsis

Use this command to deploy Kyma on a running Kubernetes cluster.

```bash
kyma alpha deploy [flags]
```

## Flags

```bash
  -a, --atomic                       Set --atomic=true to use atomic deployment, which rolls back any component that could not be installed successfully.
  -c, --components string            Path to the components file (default: "workspace/installation/resources/components.yaml") (default "workspace/installation/resources/components.yaml")
      --concurrency int              Number of parallel processes (default: 4) (default 4)
  -d, --domain string                Custom domain used for installation
  -p, --profile string               Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: "evaluation", "production".
  -s, --source string                Installation source:
                                     	- Deploy a specific release, for example: "kyma alpha deploy --source=1.17.1"
                                     	- Deploy the master branch of the Kyma repository on kyma-project.org: "kyma alpha deploy --source=master"
                                     	- Deploy a commit, for example: "kyma alpha deploy --source=34edf09a"
                                     	- Deploy a pull request, for example "kyma alpha deploy --source=PR-9486"
                                     	- Deploy the local sources: "kyma alpha deploy --source=local" (default: "master") (default "master")
      --timeout duration             Maximum time for the deployment (default: 20m0s) (default 20m0s)
      --timeout-component duration   Maximum time to deploy the component (default: 6m0s) (default 6m0s)
      --tls-crt string               TLS certificate file for the domain used for installation
      --tls-key string               TLS key file for the domain used for installation
      --value strings                Set one or more configuration values (e.g. --value component.key='the value')
  -f, --values-file strings          Path(s) to one or more JSON or YAML files with configuration values
  -w, --workspace string             Path to download Kyma sources (default: "workspace") (default "workspace")
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

