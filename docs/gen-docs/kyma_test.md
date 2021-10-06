---
title: kyma test
---

[Deprecated] Runs tests on a provisioned Kyma cluster.

## Synopsis

[DEPRECATED: the "test" command works only with Kyma 1.x.x]

Use this command to run tests on a provisioned Kyma cluster.

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
* [kyma test definitions](#kyma-test-definitions-kyma-test-definitions)	 - [Deprecated] Shows test definitions available for a provisioned Kyma cluster.
* [kyma test delete](#kyma-test-delete-kyma-test-delete)	 - [Deprecated] Deletes test suites available for a provisioned Kyma cluster.
* [kyma test list](#kyma-test-list-kyma-test-list)	 - [Deprecated] Lists test suites available for a provisioned Kyma cluster.
* [kyma test logs](#kyma-test-logs-kyma-test-logs)	 - [Deprecated] Shows the logs of tests Pods for a given test suite.
* [kyma test run](#kyma-test-run-kyma-test-run)	 - [Deprecated] Runs tests on a Kyma cluster.
* [kyma test status](#kyma-test-status-kyma-test-status)	 - [Deprecated] Shows the status of a test suite and related test executions.

