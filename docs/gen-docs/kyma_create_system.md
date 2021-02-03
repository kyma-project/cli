---
title: kyma create system
---

Creates a system on the Kyma cluster with the specified name.

## Synopsis

Use this command to create a system on the Kyma cluster.

### Detailed description

A system in Kyma is used to bind external systems and applications to the cluster and allow Kyma services and functions to communicate with them and receive events.
Once a system is created in Kyma, use the token provided by this command to allow the external system to access the Kyma cluster.

To generate a new token, rerun the same command with the `--update` flag.



```bash
kyma create system SYSTEM_NAME [flags]
```

## Options

```bash
  -n, --namespace string   Namespace to bind the system to.
  -o, --output string      Specifies the format of the command output. Supported formats: YAML, JSON.
      --timeout duration   Timeout after which CLI stops watching the installation progress. (default 2m0s)
  -u, --update             Updates an existing system and/or generates a new token for it.
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

* [kyma create](#kyma-create-kyma-create)	 - Creates resources on the Kyma cluster.

