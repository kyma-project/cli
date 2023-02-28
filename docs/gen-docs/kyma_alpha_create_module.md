---
title: kyma alpha create module
---

Creates a module bundled as an OCI image with the given OCI image name from the contents of the given path

## Synopsis

Use this command to create a Kyma module and bundle it as an OCI image.

### Detailed description

Kyma modules are individual components that can be deployed into a Kyma runtime. Modules are built and distributed as OCI container images. 
With this command, you can create such images out of a folder's contents.

This command creates a component descriptor in the descriptor path (./mod as a default) and packages all the contents on the provided path as an OCI image.
Kubebuilder projects are supported. If the path contains a kubebuilder project, it will be built and pre-defined layers will be created based on its known contents.

Alternatively, a custom (non kubebuilder) module can be created by providing a path that does not contain a kubebuilder project. In that case all the contents of the path will be bundled as a single layer.

Optionally, you can manually add additional layers with contents in other paths (see [resource flag](#flags) for more information).

Finally, if you provided a registry to which to push the artifact, the created module is validated and pushed. During the validation the default CR defined in the optional "default.yaml" file is validated against CustomResourceDefinition.
Alternatively, you can trigger an on-demand default CR validation with "--validateCR=true", in case you don't push to the registry.

To push the artifact into some registries, for example, the central docker.io registry, you have to change the OCM Component Name Mapping with the following flag: "--name-mapping=sha256-digest". This is necessary because the registry does not accept artifact URLs with more than two path segments, and such URLs are generated with the default name mapping: "urlPath". In the case of the "sha256-digest" mapping, the artifact URL contains just a sha256 digest of the full Component Name and fits the path length restrictions.



```bash
kyma alpha create module --name MODULE_NAME --version MODULE_VERSION --registry MODULE_REGISTRY [flags]
```

## Examples

```bash
Examples:
Build module my-domain/modA in version 1.2.3 and push it to a remote registry
		kyma alpha create module -n my-domain/modA --version 1.2.3 -p /path/to/module --registry https://dockerhub.com
Build module my-domain/modB in version 3.2.1 and push it to a local registry "unsigned" subfolder without tls
		kyma alpha create module -n my-domain/modB --version 3.2.1 -p /path/to/module --registry http://localhost:5001/unsigned --insecure

```

## Flags

```bash
      --channel string                     Channel to use for the module template. (default "regular")
  -c, --credentials string                 Basic authentication credentials for the given registry in the user:password format
      --default-cr string                  File containing the default custom resource of the module. If the module is a kubebuilder project, the default CR is automatically detected.
      --descriptor-version string          Schema version to use for the generated OCM descriptor. One of ocm.software/v3alpha1,v2 (default "v2")
      --insecure                           Uses an insecure connection to access the registry.
      --module-archive-path string         Specifies the path where the module artifacts are locally cached to generate the image. If the path already has a module, use the "--module-archive-version-overwrite" flag to overwrite it. (default "./mod")
      --module-archive-persistence         Uses the host filesystem instead of inmemory archiving to build the module.
      --module-archive-version-overwrite   Overwrites existing component's versions of the module. If set to false, the push is a No-Op.
  -n, --name string                        Override the module name of the kubebuilder project. If the module is not a kubebuilder project, this flag is mandatory.
      --name-mapping string                Overrides the OCM Component Name Mapping, Use: "urlPath" or "sha256-digest". (default "urlPath")
  -o, --output string                      File to write the module template if the module is uploaded to a registry. (default "template.yaml")
  -p, --path string                        Path to the module's contents. (default current directory)
      --registry string                    Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor.
      --registry-cred-selector string      Label selector to identify an externally created Secret of type "kubernetes.io/dockerconfigjson". It allows the image to be accessed in private image registries. It can be used when you push your module to a registry with authenticated access. For example, "label1=value1,label2=value2".
  -r, --resource stringArray               Add an extra resource in a new layer in the <NAME:TYPE@PATH> format. If you provide only a path, the name defaults to the last path element, and the type is set to 'helm-chart'.
      --sec-scanners-config string         Path to the file holding the security scan configuration. (default "sec-scanners-config.yaml")
      --target string                      Target to use when determining where to install the module. Use 'control-plane' or 'remote'. (default "control-plane")
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

