---
title: kyma alpha create module
---

Creates a module bundled as an OCI artifact

## Synopsis

Use this command to create a Kyma module, bundle it as an OCI artifact and optionally push it to the OCI registry.

### Detailed description

This command allows you to create a Kyma module as an OCI artifact and optionally push it to the OCI registry of your choice.
For more information about a Kyma module see the [documentation](https://github.com/kyma-project/lifecycle-manager).

This command creates a module from an existing directory containing the module's source files.
The directory must be a valid git project that is publicly available.
The command supports two types of directory layouts for the module:
- Simple: Just a directory with a valid git configuration. All the module's sources are defined in this directory.
- Kubebuilder (DEPRECATED): A directory with a valid Kubebuilder project. Module operator(s) are created using the Kubebuilder toolset.
Both simple and Kubebuilder projects require providing an explicit path to the module's project directory using the "--path" flag or invoking the command from within that directory.

### Simple mode configuration

To configure the simple mode, provide the "--module-config-file" flag with a config file path.
The module config file is a YAML file used to configure the following attributes for the module:

- name:         a string, required, the name of the module
- version:      a string, required, the version of the module
- channel:      a string, required, channel that should be used in the ModuleTemplate CR
- manifest:     a string, required, reference to the manifest, must be a relative file name
- defaultCR:    a string, optional, reference to a YAML file containing the default CR for the module, must be a relative file name
- resourceName: a string, optional, default={NAME}-{CHANNEL}, the name for the ModuleTemplate CR that will be created
- security:     a string, optional, name of the security scanners config file
- internal:     a boolean, optional, default=false, determines whether the ModuleTemplate CR should have the internal flag or not
- beta:         a boolean, optional, default=false, determines whether the ModuleTemplate CR should have the beta flag or not
- labels:       a map with string keys and values, optional, additional labels for the generated ModuleTemplate CR
- annotations:  a map with string keys and values, optional, additional annotations for the generated ModuleTemplate CR

The **manifest** and **defaultCR** paths are resolved against the module's directory, as configured with the "--path" flag.
The **manifest** file contains all the module's resources in a single, multi-document YAML file. These resources will be created in the Kyma cluster when the module is activated.
The **defaultCR** file contains a default custom resource for the module that will be installed along with the module.
The Default CR is additionally schema-validated against the Custom Resource Definition. The CRD used for the validation must exist in the set of the module's resources.

### Kubebuilder mode configuration
The Kubebuilder mode is DEPRECATED.
The Kubebuilder mode is configured automatically if the "--module-config-file" flag is not provided.

In this mode, you have to explicitly provide the module name and version using the "--name" and "--version" flags, respectively.
Some defaults, like the module manifest file location and the default CR file location, are then resolved automatically, but you can override these with the available flags.

### Modules as OCI artifacts
Modules are built and distributed as OCI artifacts. 
This command creates a component descriptor in the configured descriptor path (./mod as a default) and packages all the contents on the provided path as an OCI artifact.
The internal structure of the artifact conforms to the [Open Component Model](https://ocm.software/) scheme version 3.

If you configured the "--registry" flag, the created module is validated and pushed to the configured registry.
During the validation the **defaultCR** resource, if defined, is validated against a corresponding CustomResourceDefinition.
You can also trigger an on-demand **defaultCR** validation with "--validateCR=true", in case you don't push the module to the registry.

#### Name Mapping
To push the artifact into some registries, for example, the central docker.io registry, you have to change the OCM Component Name Mapping with the following flag: "--name-mapping=sha256-digest". This is necessary because the registry does not accept artifact URLs with more than two path segments, and such URLs are generated with the default name mapping: **urlPath**. In the case of the "sha256-digest" mapping, the artifact URL contains just a sha256 digest of the full Component Name and fits the path length restrictions. The downside of the "sha256-mapping" is that the module name is no longer visible in the artifact URL, as it contains the sha256 digest of the defined name.



```bash
kyma alpha create module [--module-config-file MODULE_CONFIG_FILE | --name MODULE_NAME --version MODULE_VERSION] [--path MODULE_DIRECTORY] [--registry MODULE_REGISTRY] [flags]
```

## Examples

```bash
Examples:
Build a simple module and push it to a remote registry
		kyma alpha create module --module-config-file=/path/to/module-config-file -path /path/to/module --registry http://localhost:5001/unsigned --insecure
Build a Kubebuilder module my-domain/modB in version 1.2.3 and push it to a remote registry
		kyma alpha create module --name my-domain/modB --version 1.2.3 --path /path/to/module --registry https://dockerhub.com
Build a Kubebuilder module my-domain/modC in version 3.2.1 and push it to a local registry "unsigned" subfolder without tls
		kyma alpha create module --name my-domain/modC --version 3.2.1 --path /path/to/module --registry http://localhost:5001/unsigned --insecure


```

## Flags

```bash
      --channel string                     Channel to use for the module template. (default "regular")
  -c, --credentials string                 Basic authentication credentials for the given registry in the user:password format
      --default-cr string                  File containing the default custom resource of the module. If the module is a kubebuilder project, the default CR is automatically detected.
      --descriptor-version string          Schema version to use for the generated OCM descriptor. One of ocm.software/v3alpha1,v2 (default "v2")
      --insecure                           Uses an insecure connection to access the registry.
      --key string                         Specifies the path where a private key is used for signing.
      --kubebuilder-project                Specifies provided module is a Kubebuilder Project.
      --module-archive-path string         Specifies the path where the module artifacts are locally cached to generate the image. If the path already has a module, use the "--module-archive-version-overwrite" flag to overwrite it. (default "./mod")
      --module-archive-persistence         Uses the host filesystem instead of in-memory archiving to build the module.
      --module-archive-version-overwrite   Overwrites existing component's versions of the module. If set to false, the push is a No-Op.
      --module-config-file string          Specifies the module configuration file
  -n, --name string                        Override the module name of the kubebuilder project. If the module is not a kubebuilder project, this flag is mandatory.
      --name-mapping string                Overrides the OCM Component Name Mapping, Use: "urlPath" or "sha256-digest". (default "urlPath")
  -o, --output string                      File to write the module template if the module is uploaded to a registry. (default "template.yaml")
  -p, --path string                        Path to the module's contents. (default current directory)
      --registry string                    Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor.
      --registry-cred-selector string      Label selector to identify an externally created Secret of type "kubernetes.io/dockerconfigjson". It allows the image to be accessed in private image registries. It can be used when you push your module to a registry with authenticated access. For example, "label1=value1,label2=value2".
  -r, --resource stringArray               Add an extra resource in a new layer in the <NAME:TYPE@PATH> format. If you provide only a path, the name defaults to the last path element, and the type is set to 'helm-chart'.
      --sec-scanners-config string         Path to the file holding the security scan configuration. (default "sec-scanners-config.yaml")
  -t, --token string                       Authentication token for the given registry (alternative to basic authentication).
      --version string                     Version of the module. This flag is mandatory.
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

