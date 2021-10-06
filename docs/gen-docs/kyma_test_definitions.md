---
title: kyma test definitions
---

[DEPRECATED] Shows test definitions available for a provisioned Kyma cluster.

## Synopsis

[DEPRECATED: the "test definitions" command works only with Kyma 1.x.x]

Use this command to list test definitions available for a provisioned Kyma cluster.

```bash
kyma test definitions [flags]
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

* [kyma test](#kyma-test-kyma-test)	 - [DEPRECATED] Runs tests on a provisioned Kyma cluster.

