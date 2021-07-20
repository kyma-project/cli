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
  -a, --agent-arg strings          One or more arguments passed to the k3s agent command on agent nodes (e.g. --agent-arg='--alsologtostderr')
      --k3d-arg strings            One or more arguments passed to the k3d provisioning command (e.g. --k3d-arg='--no-rollback')
  -k, --kube-version string        Kubernetes version of the cluster (default "1.20.7")
      --map-lb-ports stringToInt   Map ports 80 and 443 of K3D loadbalancer (e.g. --map-lb-ports=80=8080,443=8443) (default [80=8080,443=8443])
      --name string                Name of the Kyma cluster (default "kyma")
  -s, --server-arg strings         One or more arguments passed to the Kubernetes API server (e.g. --server-arg='--alsologtostderr')
      --timeout duration           Maximum time for the provisioning. If you want no timeout, enter "0". (default 5m0s)
      --workers int                Number of worker nodes (k3s agents) (default 1)
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

* [kyma alpha provision](#kyma-alpha-provision-kyma-alpha-provision)	 - Provisions a cluster for Kyma installation.

