---
title: kyma import hosts
---

Imports the hosts of exposed workloads in the system hosts file.

## Synopsis

Use this command to add the hosts of exposed workloads to the hosts file of the local system.


```bash
kyma import hosts [flags]
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

* [kyma import](kyma_import.md)	 - Imports certificates to local certificates storage or adds domains to the local host file.

