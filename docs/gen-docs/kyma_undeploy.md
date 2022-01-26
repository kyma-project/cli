---
title: kyma undeploy
---

Undeploys Kyma from a running Kubernetes cluster.

## Synopsis

Use this command to undeploy Kyma from a running Kubernetes cluster.

```bash
kyma undeploy [flags]
```

## Flags

```bash
      --component strings        Provide one or more components to undeploy (e.g. --component componentName@namespace)
  -c, --components-file string   Path to the components file (default "$HOME/.kyma/sources/installation/resources/components.yaml" or ".kyma-sources/installation/resources/components.yaml")
      --concurrency int          Set maximum number of workers to run simultaneously to deploy Kyma. (default 4)
      --delete-strategy string   Specify if only Kyma resources are deleted (system) or all resources (all) (default "system")
  -d, --domain string            Custom domain used for installation.
  -p, --profile string           Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: evaluation, production.
  -s, --source string            Source of installation to be undeployed:
                                 	- Undeploy from a specific release, for example: "kyma undeploy --source=2.0.0"
                                 	- Undeploy from a specific branch of the Kyma repository on kyma-project.org: "kyma undeploy --source=<my-branch-name>"
                                 	- Undeploy from a commit (8 characters or more), for example: "kyma undeploy --source=34edf09a"
                                 	- Undeploy from a pull request, for example "kyma undeploy --source=PR-9486"
                                 	- Undeploy from the local sources: "kyma undeploy --source=local" (default "2.0.4")
      --timeout duration         Maximum time for the deletion (default 6m0s)
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
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.

