---
title: kyma delete
---

Deletes Kyma from a running Kubernetes cluster.

## Synopsis

Use this command to delete Kyma from a running Kubernetes cluster.

```bash
kyma delete [flags]
```

## Flags

```bash
      --concurrency int              Number of parallel processes (default 4)
      --keep-crds                    Set --keep-crds=true to keep CRDs on deletion
      --timeout duration             Maximum time for the deletion (default 20m0s)
      --timeout-component duration   Maximum time to delete the component (default 6m0s)
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

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.

