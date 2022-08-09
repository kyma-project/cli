---
title: kyma alpha init module
---

Initializes an empty module with the given name.

## Synopsis

Use this command to create an empty Kyma module with the given name in the current working directory, or at some other location specified by the --path flag.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. 
With this command, you can initialize an empty module folder for the purpose of further development.

This command creates a module directory in the current working directory.
To create the module directory at a different location, use the "--path" flag.
The name of the  module directory is the same as the name of the module.
Module name must start with a letter and may only consist of alphanumeric and '.', '_', or '-' characters.
In the module directory, you'll find the template files and subdirectories corresponding to the required module structure:
    charts/       // folder containing a set of charts (each in a subfolder)
    operator/     // folder containing the operator needed to manage the module
    config.yaml   // YAML file containing the installation configuration for any chart in the module that requires custom Helm settings
	default.yaml  // YAML file containing the default CR needed to start the module installation.
    README.md     // document explaining the module format, how it translates to OCI images, and how to develop one (can be mostly empty at the beginning)


```bash
kyma alpha init module --name=MODULE_NAME [--path=PARENT_DIR] [flags]
```

## Flags

```bash
      --name string   Specifies the module name
      --path string   Specifies the path where the module directory is created. The path must exist
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

* [kyma alpha init](kyma_alpha_init.md)	 - Initializes resources locally.

