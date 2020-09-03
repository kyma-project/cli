---
title: kyma template function
---

Renders Function resources from the local workspace.

## Synopsis

Use this command to print and validate all Function-related resources generated from the sources in your local workspace.
Use the flags to specify the location of the source files or their output format.

```bash
kyma template function [flags]
```

## Options

```bash
  -d, --dir string      Full path to the directory where you want to save the project
  -o, --output string   Output format. Use one of:
                        - json
                        - yaml (default "json")
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

* [kyma template](#kyma-template-kyma-template)	 - Renders resources from the local workspace.

