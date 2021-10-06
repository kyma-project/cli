---
title: kyma console
---

[DEPRECATED] Opens the Kyma Console in a web browser.

## Synopsis

[DEPRECATED: the "console" command works only with Kyma 1.x.x. Please use the "dashboard" command with Kyma 2.x.x]
		
Use this command to open the Kyma Console in a web browser.

```bash
kyma console [flags]
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

