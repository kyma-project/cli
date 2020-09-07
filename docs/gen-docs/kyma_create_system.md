---
title: kyma create system
---

Creates a system on the Kyma cluster.

## Synopsis

Use this command to create a system on a Kyma cluster.

### Detailed description

A system in Kyma is used to bind external systems and applications into the cluster and allow Kyma services and functions to receive events and communicate with them.
Once a system is created in kyma, use the token provided by this command to grant access to the external system into the kyma cluster.

To generate a new token, rerun the same command with the --update flag.



```bash
kyma create system [flags]
```

## Options

```bash
  -n, --namespace string   Namespace to bind the system to.
  -o, --output string      Specify the format of the output of the command. Supported formats: yaml.
      --timeout duration   Time-out after which CLI stops watching the installation progress. (default 2m0s)
  -u, --update             Update an existing system and/or generate a new token for it.
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

* [kyma create](#kyma-create-kyma-create)	 - Creates resources on the Kyma cluster

