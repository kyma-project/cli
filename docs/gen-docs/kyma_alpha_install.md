---
title: kyma alpha install
---

Installs Kyma on a running Kubernetes cluster.

## Synopsis

Use this command to install Kyma on a running Kubernetes cluster.

```bash
kyma alpha install [flags]
```

## Options

```bash
  -c, --components string   Path to a YAML file with component list to override.
  -o, --overrides string    Path to a YAML file with parameters to override.
  -r, --resources string    Path to Kyma resources folder.
```

## Options inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (e.g. no dialog prompts) and ensures that logs are formatted properly in log files (e.g. no spinners for CLI steps).
  -h, --help                Displays help for the command.
      --kubeconfig string   Specifies the path to the kubeconfig file. By default, Kyma CLI uses the KUBECONFIG environment variable or "/$HOME/.kube/config" if the variable is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma alpha](#kyma-alpha-kyma-alpha)	 - Executes the commands in alpha testing stage.

