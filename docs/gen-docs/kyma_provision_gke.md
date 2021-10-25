---
title: kyma provision gke
---

Provisions a Google Kubernetes Engine (GKE) cluster on Google Cloud Platform (GCP).

## Synopsis

Use this command to provision a GKE cluster on GCP for Kyma installation. Use the flags to specify cluster details.
NOTE: To access the provisioned cluster, make sure you get authenticated by Google Cloud SDK. To do so,run `gcloud auth application-default login` and log in with your Google Cloud credentials.

```bash
kyma provision gke [flags]
```

## Flags

```bash
      --attempts uint         Maximum number of attempts to provision the cluster. (default 3)
  -c, --credentials string    Path to the GCP service account key file. (required)
      --disk-size int         Disk size (in GB) of the cluster. (default 50)
  -k, --kube-version string   Kubernetes version of the cluster. (default "1.19")
  -l, --location string       Region (e.g. europe-west3) or zone (e.g. europe-west3-a) of the cluster. (default "europe-west3-a")
  -n, --name string           Name of the GKE cluster to provision. (required)
      --nodes int             Number of cluster nodes. (default 3)
  -p, --project string        Name of the GCP Project where you provision the GKE cluster. (required)
  -t, --type string           Machine type used for the cluster. (default "n1-standard-4")
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma provision](#kyma-provision-kyma-provision)	 - Provisions a cluster for Kyma installation.

