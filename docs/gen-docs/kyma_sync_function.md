---
title: kyma sync function
---

Synchronizes the local resources for your Function.

## Synopsis

Use this command to download the Function's code and dependencies from the cluster to create or update these resources in your local workspace.
Use the flags to specify the name of your Function, the Namespace, or the location for your project.

```bash
kyma sync function [flags]
```

## Options

```bash
      --name string        Function name.
  -n, --namespace string   Namespace from which you want to sync the Function.
  -o, --output string      Full path to the directory where you want to save the project.
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

* [kyma sync](#kyma-sync-kyma-sync)	 - Synchronizes the local resources for your Function.

