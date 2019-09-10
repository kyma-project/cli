## kyma test status

Shows the status of a test suite and it's related test executions

### Synopsis

Use this command to display the status of a test suite and related test executions.

If you don't provide any arguments, the status of all test suites will be printed.

To print the status of all test suites, run:

    kyma test status

To print the status of specific test cases, run:

	kyma test status testSuiteOne testSuiteTwo

```
kyma test status <test-suite-1> <test-suite-2> ... <test-suite-N> [flags]
```

### Options

```
  -h, --help            Displays help for the command.
  -o, --output string   Output format. One of: json|yaml|wide
```

### Options inherited from parent commands

```
      --kubeconfig string   Specifies the path to the kubeconfig file. (default "/Users/user/.kube/config")
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

### SEE ALSO

* [kyma test](kyma_test.md)	 - Run tests on a provisioned Kyma cluster.

