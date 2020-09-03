---
title: kyma apply function
---

Applies local resources for your Function to the Kyma cluster.

## Synopsis

Use this command to apply the local sources of your Function's code and dependencies to the Kyma cluster. 
Use the flags to specify the desired location for the source files or run the command to validate and print the output resources.

```bash
kyma apply function [flags]
```

## Options

```bash
      --dry-run           Validated list of objects to be created from sources.
  -f, --filename string   Full path to the config file.
      --onerror string    Flag used to define reaction to the error. Use one of:
                          - nothing
                          - purge (default "nothing")
  -o, --output string     Flag used to define output of the command. Use one of:
                          - text
                          - json
                          - yaml
                          - none (default "text")
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

* [kyma apply](#kyma-apply-kyma-apply)	 - Applies local resources to the Kyma cluster.

