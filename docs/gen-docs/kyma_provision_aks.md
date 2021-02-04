---
title: kyma provision aks
---

Provisions an Azure Kubernetes Service (AKS) cluster on Azure.

## Synopsis

Use this command to provision an AKS cluster on Azure for Kyma installation. Use the flags to specify cluster details. 
	NOTE: To provision and access the provisioned cluster, make sure you get authenticated by using the Azure CLI. To do so,run `az login` and log in with your Azure credentials.

```bash
kyma provision aks [flags]
```

## Options

```bash
      --attempts uint         Maximum number of attempts to provision the cluster. (default 3)
  -c, --credentials string    Path to the TOML file containing the Azure Subscription ID (SUBSCRIPTION_ID), Tenant ID (TENANT_ID), Client ID (CLIENT_ID) and Client Secret (CLIENT_SECRET). (required)
      --disk-size int         Disk size (in GB) of the cluster. (default 50)
  -k, --kube-version string   Kubernetes version of the cluster. (default "1.16.15")
  -l, --location string       Location of the cluster. (default "westeurope")
  -n, --name string           Name of the AKS cluster to provision. (required)
      --nodes int             Number of cluster nodes. (default 3)
  -p, --project string        Name of the Azure Resource Group where you provision the AKS cluster. (required)
  -t, --type string           Machine type used for the cluster. (default "Standard_D4_v3")
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

* [kyma provision](#kyma-provision-kyma-provision)	 - Provisions a cluster for Kyma installation.

