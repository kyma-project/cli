---
title: kyma init function
---

Creates local resources for your Function.

## Synopsis

Use this command to create the local workspace with the default structure of your Function's code and dependencies. Update this configuration to your references and apply it to a Kyma cluster. 
Use the flags to specify the initial configuration for your Function or to choose the location for your project.

```bash
kyma init function [flags]
```

## Options

```bash
  -d, --dir string         Full path to the directory where you want to save the project.
      --name string        Function name. (default "first-function")
      --namespace string   Namespace to which you want to apply your Function. (default "default")
  -r, --runtime string     Flag used to define the environment for running you Function. Use one of:
                           	- nodejs12
                           	- nodejs10
                           	- python38 (default "nodejs12")
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

* [kyma init](#kyma-init-kyma-init)	 - Creates local resources for your project.

