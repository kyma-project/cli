---
title: kyma dashboard
---

Runs the Kyma dashboard locally and opens it directly in a web browser.

## Synopsis

Use this command to run the Kyma dashboard locally in a docker container and open it directly in a web browser. This command only works with a local installation of Kyma.

```bash
kyma dashboard [flags]
```

## Flags

```bash
      --container-name string   Specify the name of the local container. (default "kyma-dashboard")
  -p, --port string             Specify the port on which the local dashboard will be exposed. (default "3001")
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

* [kyma](kyma.md)	 - Controls a Kyma cluster.

