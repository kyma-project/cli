---
title: kyma alpha create scaffold
---

Generates necessary files required for module creation

## Synopsis

Scaffold generates the necessary files for creating a new module in Kyma. This includes setting up 
a basic directory structure and creating default files based on the provided flags.
The command generates the following files:
 - Module Config - module-config.yaml (always generated)
 - Manifest - template-operate.yaml (generated when the "--gen-manifest" flag is set)
 - Security Scanners Config - sec-scanners-config.yaml (generated when the "--gen-sec-config" flag is set)
 - Default CR - config/samples/operator.kyma-project.io_v1alpha1_sample.yaml (generated when the "--gen-default-cr" is flag set)

You must specify the required fields of the module config using the following CLI arguments:
--module-name [NAME]
--module-version [VERSION]
--module-channel [CHANNEL]
--module-manifest-path [MANIFEST-PATH] (cannot be used with the "--gen-manifest" flag)

**NOTE:**: If the required fields aren't provided, the module-config.yaml is not ready to use out-of-the-box. You must manually edit the file to make it usable.
Also, edit the sec-scanners-config.yaml to be able to use it.

The command is designed to streamline the module creation process in Kyma, making it easier and more 
efficient for developers to get started with new modules. It supports customization through various flags, 
allowing for a tailored scaffolding experience according to the specific needs of the module being created.

```bash
kyma alpha create scaffold [--module-name MODULE_NAME --module-version MODULE_VERSION --module-channel CHANNEL --module-manifest] [--directory MODULE_DIRECTORY] [flags]
```

## Examples

```bash
Examples:
Generate a simple scaffold for a module
		kyma alpha create scaffold --module-name=template-operator --module-version=1.0.0 --module-channel=regular --module-manifest-path=./template-operator.yaml
Generate a scaffold with manifest file, default CR, and security config for a module
		kyma alpha create scaffold --module-name=template-operator --module-version=1.0.0 --module-channel=regular --gen-manifest --gen-sec-config --gen-default-cr

```

## Flags

```bash
  -d, --directory string              Specifies the directory where the scaffolding shall be generated (default "./")
      --gen-default-cr                Specifies if a default CR should be generated
      --gen-manifest                  Specifies if manifest file should be generated
      --gen-sec-config                Specifies if security config should be generated
      --module-channel string         Specifies the module channel in the generated module config file
      --module-manifest-path string   Specifies the module manifest filepath in the generated module config file
      --module-name string            Specifies the module name in the generated module config file
      --module-version string         Specifies the module version in the generated module config file
  -o, --overwrite                     Specifies if the scaffold overwrites existing files
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

* [kyma alpha create](kyma_alpha_create.md)	 - Creates resources on the Kyma cluster.

