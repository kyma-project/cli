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

## Flags

```bash
      --base-dir string                 A directory in the repository containing the Function's sources (default "/")
  -d, --dir string                      Full path to the directory where you want to save the project.
      --name string                     Function name.
      --namespace string                Namespace to which you want to apply your Function.
      --reference string                Commit hash or branch name (default "main")
      --repository-name string          The name of the Git repository to be created
  -r, --runtime string                  Flag used to define the environment for running your Function. Use one of these options:
                                        	- nodejs18 
                                        	- nodejs20
                                        	- python39 (deprecated)
                                        	- python312 (default "nodejs18")
      --runtime-image-override string   Set custom runtime image base.
      --schema-version string           Version of the config API. (default "v0")
      --url string                      Git repository URL
      --vscode                          Generate VS Code settings containing config.yaml JSON schema for autocompletion (see "kyma get schema -h" for more info)
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner).
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma init](kyma_init.md)	 - Creates local resources for your project.

