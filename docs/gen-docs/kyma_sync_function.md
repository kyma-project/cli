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

## Flags

```bash
  -d, --dir string              Full path to the directory where you want to save the project.
  -n, --namespace string        Namespace from which you want to sync the Function.
      --schema-version string   Version of the config API. (default "v0")
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

* [kyma sync](kyma_sync.md)	 - Synchronizes the local resources for your Function.

