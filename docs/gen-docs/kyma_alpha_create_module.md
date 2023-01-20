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

Finally, if you provided a registry to which to push the artifact, the created module is validated and pushed. For example, the default CR defined in the \"default.yaml\" file is validated against CustomResourceDefinition.

Alternatively, if you don't push to registry, you can trigger an on-demand validation with "--validateCR=true".


```bash
kyma alpha create module [flags]
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
      --channel string                  Channel to use for the module template. (default "regular")
      --clean                           Remove the mod-path folder and all its contents at the end.
  -c, --credentials string              Basic authentication credentials for the given registry in the format user:password
      --default-cr string               File containing the default custom resource of the module. If the module is a kubebuilder project, the default CR will be automatically detected.
      --insecure                        Use an insecure connection to access the registry.
      --mod-cache string                Specifies the path where the module artifacts are locally cached to generate the image. If the path already has a module, use the overwrite flag to overwrite it. (default "./mod")
  -n, --name string                     Override the module name of the kubebuilder project. If the module is not a kubebuilder project, this flag is mandatory.
  -o, --output string                   File to which to output the module template if the module is uploaded to a registry (default "template.yaml")
  -w, --overwrite                       overwrites the existing mod-path directory if it exists
  -p, --path string                     Path to the module contents. (default current directory)
      --registry string                 Repository context url for module to upload. The repository url will be automatically added to the repository contexts in the module
      --registry-cred-selector string   label selector to identify a secret of type kubernetes.io/dockerconfigjson (that needs to be created externally) which allows the image to be accessed in private image registries. This can be used if you push your module to a registry with authenticated access. Example: "label1=value1,label2=value2"
  -r, --resource stringArray            Add an extra resource in a new layer with format <NAME:TYPE@PATH>. It is also possible to provide only a path; name will default to the last path element and type to 'helm-chart'
  -t, --token string                    Authentication token for the given registry (alternative to basic authentication).
      --version string                  Version of the module. This flag is mandatory.
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

* [kyma alpha create](kyma_alpha_create.md)	 - Creates resources on the Kyma cluster.

