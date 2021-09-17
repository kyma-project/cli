---
title: kyma alpha deploy
---

Deploys Kyma on a running Kubernetes cluster.

## Synopsis

Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.

```bash
kyma alpha deploy [flags]
```

## Flags

```bash
      --component strings        Provide one or more components to deploy (e.g. --component componentName@namespace)
  -c, --components-file string   Path to the components file (default "$HOME/.kyma/sources/installation/resources/components.yaml" or ".kyma-sources/installation/resources/components.yaml")
  -d, --domain string            Custom domain used for installation.
  -p, --profile string           Kyma deployment profile. If not specified, Kyma uses its default configuration. The supported profiles are: evaluation, production.
      --tls-crt string           TLS certificate file for the domain used for installation.
      --tls-key string           TLS key file for the domain used for installation.
      --value strings            Set configuration values. Can specify one or more values, also as a comma-separated list (e.g. --value component.a='1' --value component.b='2' or --value component.a='1',component.b='2').
  -f, --values-file strings      Path(s) to one or more JSON or YAML files with configuration values.
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

