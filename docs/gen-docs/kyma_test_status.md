---
title: kyma test status
---

Shows the status of a test suite and related test executions.

## Synopsis

Use this command to display the status of a test suite and related test executions.

If you don't provide any arguments, the status of all test suites will be printed.
To print the status of all test suites, run `kyma test status`.
To print the status of specific test cases, run `kyma test status testSuiteOne testSuiteTwo`.


```bash
kyma test status <test-suite-1> <test-suite-2> ... <test-suite-N> [flags]
```

## Options

```bash
  -o, --output string   Output format. One of: json|yaml|wide|junit
```

## Options inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems.
  -h, --help                Displays help for the command.
      --kubeconfig string   Specifies the path to the kubeconfig file. By default, Kyma CLI uses the KUBECONFIG environment variable or "/$HOME/.kube/config" if the variable is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma test](#kyma-test-kyma-test)	 - Runs tests on a provisioned Kyma cluster.

