---
title: kyma version
---

Displays the version of Kyma CLI and the connected Kyma cluster.

## Synopsis

Use this command to print the version of Kyma CLI and the version of the Kyma cluster the current kubeconfig points to.


```bash
kyma version [flags]
```

## Flags

```bash
  -c, --client   Client version only (no server required)
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                See help for the command
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.

