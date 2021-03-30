---
title: kyma test
---

Runs tests on a provisioned Kyma cluster.

## Synopsis

Use this command to run tests on a provisioned Kyma cluster.

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                See help for the command
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back to "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.
* [kyma test definitions](#kyma-test-definitions-kyma-test-definitions)	 - Shows test definitions available for a provisioned Kyma cluster.
* [kyma test delete](#kyma-test-delete-kyma-test-delete)	 - Deletes test suites available for a provisioned Kyma cluster.
* [kyma test list](#kyma-test-list-kyma-test-list)	 - Lists test suites available for a provisioned Kyma cluster.
* [kyma test logs](#kyma-test-logs-kyma-test-logs)	 - Shows the logs of tests Pods for a given test suite.
* [kyma test run](#kyma-test-run-kyma-test-run)	 - Runs tests on a Kyma cluster.
* [kyma test status](#kyma-test-status-kyma-test-status)	 - Shows the status of a test suite and related test executions.
