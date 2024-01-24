---
title: kyma deploy
---

Deploys Kyma on a running Kubernetes cluster.

## Synopsis

Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.

```bash
kyma deploy [flags]
```

## Flags

```bash
      --component stringArray    Provide one or more components to deploy, for example:
                                 	- With short-hand notation: "--component name@namespace"
                                 	- With verbose JSON structure "--component '{"name": "componentName","namespace": "componentNamespace","url": "componentUrl","version": "1.2.3"}'
  -c, --components-file string   Path to the components file (default "$HOME/.kyma/sources/installation/resources/components.yaml" or ".kyma-sources/installation/resources/components.yaml")
      --concurrency int          Set maximum number of workers to run simultaneously to deploy Kyma. (default 4)
  -d, --domain string            Custom domain used for installation.
      --dry-run                  Alpha feature: Renders the Kubernetes manifests without actually applying them. The generated resources are not sufficient to apply Kyma to a cluster, because components having custom installation routines (such as Istio) are not included.
  -p, --profile string           Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: evaluation, production.
  -s, --source string            Installation source:
                                 	- Deploy a specific release, for example: "kyma deploy --source=2.0.0"
                                 	- Deploy a specific branch of the Kyma repository on kyma-project.org: "kyma deploy --source=<my-branch-name>"
                                 	- Deploy a commit (8 characters or more), for example: "kyma deploy --source=34edf09a"
                                 	- Deploy a pull request, for example "kyma deploy --source=PR-9486"
                                 	- Deploy the local sources: "kyma deploy --source=local" (default "2.20.0")
  -t, --timeout duration         Maximum time for the deployment. (default 20m0s)
      --tls-crt string           TLS certificate file for the domain used for installation.
      --tls-key string           TLS key file for the domain used for installation.
      --value strings            Set configuration values. Can specify one or more values, also as a comma-separated list (e.g. --value component.a='1' --value component.b='2' or --value component.a='1',component.b='2').
  -f, --values-file strings      Path(s) to one or more JSON or YAML files with configuration values.
  -w, --workspace string         Path to download Kyma sources (default "$HOME/.kyma/sources" or ".kyma-sources")
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner).
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](kyma.md)	 - Controls a Kyma cluster.

