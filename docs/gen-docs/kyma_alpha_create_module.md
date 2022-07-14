---
title: kyma alpha create module
---

Creates a module bundled as an OCI image from the contents of the given path

## Synopsis

Use this command to create a Kyma module.

### Detailed description

Kyma modules are individual components that can be deplyed into a Kyma runtime. Modules are built and distributed as OCI continer images. 
This command provides the means to create such images out of a folder's contents.

This command will create a component descriptor in the descriptor path (./mod as a default) and package all the contents on the provided content path as a single layer.
Optionally, additional layers can be added with contents in other paths.

Finally, if a registry is provided, the created module will be pushed.


```bash
kyma alpha create module MODULE_NAME MODULE_VERSION <CONTENT_PATH> [flags]
```

## Flags

```bash
      --clean                  Remove the mod-path folder and all its contents at the end.
  -c, --credentials string     Basic authentication credentials for the given registry in the format user:password
      --insecure               Use an insecure connection to access the registry.
      --mod-path string        Specifies the path where the component descriptor and module packaging will be stored. If the path already has a descriptor use the overwrite flag to overwrite it (default "./mod")
      --name-mapping string    Repository context name mapping. Possible values are: empty, urlPath or sha256-digest
  -w, --overwrite              overwrites the existing mod-path directory if it exists
      --registry string        Repository context url for component to upload. The repository url will be automatically added to the repository contexts in the module
  -r, --resource stringArray   Add an extra resource in a new layer with format <NAME:TYPE@PATH>. It is also possible to provide only a path; name will default to the last path element and type to 'helm-chart'
  -t, --token string           Authentication token for the given registry (alternative to basic authentication).
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

