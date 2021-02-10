---
title: kyma alpha provision k3s
---

Provisions a Kubernetes cluster based on k3s.

## Synopsis

Use this command to provision a k3s-based Kubernetes cluster for Kyma installation.

```bash
kyma alpha provision k3s [flags]
```

## Options

```bash
      --name string           Name of the Kyma cluster. (default "kyma")
  -s, --server-args strings   Arguments passed to the Kubernetes server (e.g. --server-args='--alsologtostderr').
      --timeout duration      Maximum time in minutes during which the provisioning takes place, where "0" means "infinite". (default 5m0s)
      --workers int           Number of worker nodes. (default 1)
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

* [kyma alpha provision](#kyma-alpha-provision-kyma-alpha-provision)	 - Provisions a cluster for Kyma installation.

