---
title: kyma run function
---

Run functions locally.

## Synopsis

Use this command to run function in docker from local sources.

```bash
kyma run function [flags]
```

## Options

```bash
      --containerName string   The name of the created container.
  -d, --detach                 Change this flag to true if you don't want to follow the container logs after run'.
  -e, --env stringArray        The system environments witch which the container will be run.
  -f, --filename string        Full path to the config file.
      --imageName string       Full name with tag of the container.
  -p, --port string            The port on which the container will be exposed. (default "8080")
  -t, --timeout duration       Maximum time during which the local resources are being built, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
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

