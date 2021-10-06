---
title: kyma test status
---

[DEPRECATED] Shows the status of a test suite and related test executions.

## Synopsis

[DEPRECATED: the "test status" command works only with Kyma 1.x.x]
		
Use this command to display the status of a test suite and related test executions.

If you don't provide any arguments, the status of all test suites will be printed.
To print the status of all test suites, run `kyma test status`.
To print the status of specific test cases, run `kyma test status testSuiteOne testSuiteTwo`.

```bash
kyma test status <test-suite-1> <test-suite-2> ... <test-suite-N> [flags]
```

## Flags

```bash
  -o, --output string   Output format. One of: json|yaml|wide|junit
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

