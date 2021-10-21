---
title: kyma run function
---

Runs Functions locally.

## Synopsis

Use this command to run a Function in Docker from local sources.

```bash
kyma run function [flags]
```

## Flags

```bash
      --container-name string   The name of the created container.
      --debug                   Change this flag to "true" if you want to expose port 9229 for remote debugging.
      --detach                  Change this flag to "true" if you don't want to follow the container logs after running the Function.
  -f, --filename string         Full path to the config file.
      --hot-deploy              Change this flag to "true" if you want to start a Function in Hot Deploy mode.
  -p, --port string             The port on which the container will be exposed. (default "8080")
  -d, --source-dir string       Full path to the folder with the source code.
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

* [kyma run](#kyma-run-kyma-run)	 - Runs resources.

