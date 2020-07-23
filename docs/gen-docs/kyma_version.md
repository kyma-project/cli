---
title: kyma version
---

Displays the version of Kyma CLI and the connected Kyma cluster.

## Synopsis

Use this command to print the version of Kyma CLI and the version of the Kyma cluster the current kubeconfig points to.


```bash
kyma version [flags]
```

## Options

```bash
  -c, --client   Client version only (no server required)
```

## Options inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems.
  -h, --help                Displays help for the command.
      --kubeconfig string   Specifies the path to the kubeconfig file. By default, Kyma CLI uses the KUBECONFIG environment variable or "/$HOME/.kube/config" if the variable is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.

