---
title: kyma test logs
---

[Deprecated] Shows the logs of tests Pods for a given test suite.

## Synopsis

[DEPRECATED: the "test logs" command works only with Kyma 1.x.x]

Use this command to display logs of a test executed for a given test suite. By default, the command displays logs for failed tests, but you can change this behavior using the "test-status" flag. 

To print the status of specific test cases, run `kyma test logs testSuiteOne testSuiteTwo`.
Provide at least one test suite name.

```bash
kyma test logs <test-suite-1> <test-suite-2> ... <test-suite-N> [flags]
```

## Flags

```bash
      --ignored-containers strings   Container names which are ignored when fetching logs from testing Pods. Takes comma-separated list. (default [istio-init,istio-proxy,manager])
      --test-status string           Displays logs coming only from testing Pods with a given status. (default "Failed")
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

* [kyma test](#kyma-test-kyma-test)	 - [Deprecated] Runs tests on a provisioned Kyma cluster.

