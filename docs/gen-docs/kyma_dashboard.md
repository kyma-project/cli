---
title: kyma dashboard
---

Opens the Kyma dashboard in a web browser.

## Synopsis

Use this command to open the Kyma dashboard in a web browser.

```bash
kyma dashboard [flags]
```

## Flags

```bash
      --container-name string   Specify the name of the local container. (default "busola")
  -d, --detach                  Change this flag to "true" if you don't want to follow the logs of the local container.
  -p, --port string             Specify the port on which the local dashboard will be exposed. (default "3001")
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

