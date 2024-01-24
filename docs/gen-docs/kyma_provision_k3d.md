---
title: kyma provision k3d
---

Provisions a Kubernetes cluster based on k3d v5.

## Synopsis

Use this command to provision a k3d-based Kubernetes cluster for Kyma installation.

```bash
kyma provision k3d [flags]
```

## Flags

```bash
      --k3d-arg strings            One or more arguments passed to the k3d provisioning command (e.g. --k3d-arg='--no-rollback')
      --k3d-registry-arg strings   One or more arguments passed to the k3d registry create command (e.g. --k3d-registry-arg='--default-network podman')
  -s, --k3s-arg strings            One or more arguments passed from k3d to the k3s command (format: ARG@NODEFILTER[;@NODEFILTER])
  -k, --kube-version string        Kubernetes version of the cluster (default "1.27.9")
      --name string                Name of the Kyma cluster (default "kyma")
  -p, --port strings               Map ports 80 and 443 of K3D loadbalancer (e.g. -p 80:80@loadbalancer -p 443:443@loadbalancer) (default [80:80@loadbalancer,443:443@loadbalancer])
      --registry-port string       Specify the port on which the k3d registry will be exposed (default "5001")
      --registry-use strings       Connect to one or more k3d-managed registries. Kyma automatically creates a registry for Serverless images.
      --timeout duration           Maximum time for the provisioning. If you want no timeout, enter "0". (default 5m0s)
      --workers int                Number of worker nodes (k3d agents) (default 1)
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

* [kyma provision](kyma_provision.md)	 - Provisions a cluster for Kyma installation.

