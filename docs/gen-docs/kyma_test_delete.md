---
title: kyma test delete
---

Deletes test suites available for a provisioned Kyma cluster.

## Synopsis

Use this command to delete test suites available for a provisioned Kyma cluster.

Provide at least one test suite name.

```bash
kyma test delete <test-suite-1> <test-suite-2> ... <test-suite-N> [flags]
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

