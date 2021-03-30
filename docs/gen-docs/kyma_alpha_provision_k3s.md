---
title: kyma alpha provision k3s
---

Provisions a Kubernetes cluster based on k3s.

## Synopsis

Use this command to provision a k3s-based Kubernetes cluster for Kyma installation.

```bash
kyma alpha provision k3s [flags]
```

## Flags

```bash
      --name string          Name of the Kyma cluster (default: "kyma") (default "kyma")
  -s, --server-arg strings   One or more arguments passed to the Kubernetes API server (e.g. --server-arg='--alsologtostderr')
      --timeout duration     Maximum time for the provisioning (default: 5m0s). If you want no timeout, enter "0". (default 5m0s)
      --workers int          Number of worker nodes (k3s agents), default: 1 (default 1)
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                See help for the command
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back to "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma alpha provision](#kyma-alpha-provision-kyma-alpha-provision)	 - Provisions a cluster for Kyma installation.
