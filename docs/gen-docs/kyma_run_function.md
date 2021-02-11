---
title: kyma run function
---

Runs Functions locally.

## Synopsis

Use this command to run a Function in Docker from local sources

```bash
kyma run function [flags]
```

## Options

```bash
      --containerName string   The name of the created container.
      --detach                 Change this flag to "true" if you don't want to follow the container logs after running the Function.
  -f, --filename string        Full path to the config file.
  -p, --port string            The port on which the container will be exposed. (default "8080")
  -d, --sourceDir string       Full path to the folder with the source code.
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

* [kyma run](#kyma-run-kyma-run)	 - 

