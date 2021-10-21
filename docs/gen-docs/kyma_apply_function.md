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

## Flags

```bash
      --dry-run            Validated list of objects to be created from sources.
  -f, --filename string    Full path to the config file.
      --onerror value      Flag used to define the Kyma CLI's reaction to an error when applying resources to the cluster. Use one of these options: 
                           - nothing
                           - purge (default nothing)
  -o, --output value       Flag used to define the command output format. Use one of these options:
                           - text
                           - json
                           - yaml
                           - none (default text)
  -t, --timeout duration   Maximum time during which the local resources are being applied, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
  -w, --watch              Flag used to watch resources applied to the cluster to make sure that everything is applied in the correct order.
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma apply](#kyma-apply-kyma-apply)	 - Applies local resources to the Kyma cluster.

