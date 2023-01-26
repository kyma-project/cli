---
title: kyma provision gardener
---

Provisions a cluster using Gardener on GCP, Azure, or AWS.

## Synopsis

Provisions a cluster using Gardener on GCP, Azure, or AWS.

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
* [kyma provision gardener aws](kyma_provision_gardener_aws.md)	 - Provisions a Kubernetes cluster using Gardener on Amazon Web Services (AWS).
* [kyma provision gardener az](kyma_provision_gardener_az.md)	 - Provisions a Kubernetes cluster using Gardener on Azure.
* [kyma provision gardener gcp](kyma_provision_gardener_gcp.md)	 - Provisions a Kubernetes cluster using Gardener on Google Cloud Platform (GCP).

