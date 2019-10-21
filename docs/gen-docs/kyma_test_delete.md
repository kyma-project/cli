## kyma test delete

Deletes test suites available for a provisioned Kyma cluster.

### Synopsis

Use this command to delete test suites available for a provisioned Kyma cluster.

Provide at least one test suite name.

```
kyma test delete <test-suite-1> <test-suite-2> ... <test-suite-N> [flags]
```

### Options

```
  -h, --help   Displays help for the command.
```

### Options inherited from parent commands

```
      --kubeconfig string   Specifies the path to the kubeconfig file. Use the default KUBECONFIG environment variable or "/$HOME/.kube/config" if KUBECONFIG is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

### SEE ALSO

* [kyma test](kyma_test.md)	 - Run tests on a provisioned Kyma cluster.

