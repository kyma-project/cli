---
title: kyma alpha init module
---

Initializes an empty module with the given name in the given parent directory.

## Synopsis

Use this command to create an empty Kyma module with the given name in a correspondingly named subdirectory of provided parent directory.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. 
With this command, you can initialize an empty module folder for the purpose of further development.

This command creates a directory with a given name in the target directory.
In this directory, you'll find the template files and subdirectories corresponding to the required module structure:
    charts/       // folder containing a set of charts (each in a subfolder)
    crds/         // folder containing all CRDs required by the module
    operator/     // folder containing the operator needed to manage the module
    profiles/     // folder containing all profile settings
    channels/     // folder containing all channel settings
    config.yaml   // YAML file containing installation configuration for any chart in the module that requires custom Helm settings
    README.md     // document explaining the module format, how it translates to OCI images, and how to develop one (can be mostly empty at the beginning)


```bash
kyma alpha init module MODULE_NAME <PARENT_DIR> [flags]
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

