---
title: kyma alpha version
---

Displays the version of Kyma CLI and of the connected Kyma cluster.

## Synopsis

Use this command to print the version of Kyma CLI and the version of the Kyma cluster the current kubeconfig points to.


```bash
kyma alpha version [flags]
```

## Options

```bash
  -c, --client   Client version only (no server required)
```

## Options inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Displays help for the command.
      --kubeconfig string   Specifies the path to the kubeconfig file. By default, Kyma CLI uses the KUBECONFIG environment variable or "/$HOME/.kube/config" if the variable is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma alpha](#kyma-alpha-kyma-alpha)	 - Executes the commands in the alpha testing stage.

