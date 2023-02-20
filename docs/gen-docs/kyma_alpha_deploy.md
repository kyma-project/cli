---
title: kyma alpha deploy
---

Deploys Kyma on a running Kubernetes cluster.

## Synopsis

Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.

```bash
kyma alpha deploy [flags]
```

## Examples

```bash

- Deploy the latest version of the Lifecycle Manager for trying out Modules: "kyma deploy -k https://github.com/kyma-project/lifecycle-manager/config/default -with-wildcard-permissions"
- Deploy the main branch of Lifecycle Manager: "kyma deploy -k https://github.com/kyma-project/lifecycle-manager/config/default@main"
- Deploy a local version of Lifecycle Manager: "kyma deploy -k /path/to/repo/lifecycle-manager/config/default"
```

## Flags

```bash
      --cert-manager string         Installs cert-manager from the specified static version. an empty string skips the installation. (default "v1.11.0")
  -c, --channel string              Select which channel to deploy from. (default "regular")
      --dry-run                     Renders the Kubernetes manifests without actually applying them.
  -k, --kustomization stringArray   Provide one or more kustomizations to deploy. 
                                    Each occurrence of the flag accepts a URL with an optional reference (commit, branch, or release) in the format URL@ref or a local path to the directory of the kustomization file.
                                    Defaults to deploying Lifecycle Manager and Module Manager from GitHub main branch. (default [https://github.com/kyma-project/lifecycle-manager/config/default])
      --kyma-cr string              Provide a custom Kyma CR file for the deployment.
      --lifecycle-manager string    Installs lifecycle-manager with the specified image:
                                    - Use "my-registry.org/lifecycle-manager:my-tag"" to use a custom version of lifecycle-manager.
                                    - Use "europe-docker.pkg.dev/kyma-project/prod/lifecycle-manager@sha256:cb74b29cfe80c639c9ee9..." to use a custom version of lifecycle-manager with a digest.
                                    - Use a tag like "v20230220-7b8e9515" to override the default tag used, it will then use "eu.gcr.io/kyma-project/lifecycle-manager:v20230220-7b8e9515". (default "eu.gcr.io/kyma-project/lifecycle-manager:latest")
  -m, --module stringArray          Provide one or more modules to activate after the deployment is finished. Example: "--module name@namespace" (namespace is optional).
  -f, --modules-file string         Path to file containing a list of modules.
  -n, --namespace string            The Namespace to deploy the the Kyma custom resource in. (default "kyma-system")
      --open-dashboard              Opens the Busola Dashboard at startup. Only works when a graphical interface is available and when running in interactive mode
      --template stringArray        Provide one or more module templates to deploy.
                                    WARNING: This is a temporary flag for development and will be removed soon.
  -t, --timeout duration            Maximum time for the deployment. (default 20m0s)
      --wildcard-permissions        Creates a wildcard cluster-role to allow for easy local installation permissions of lifecycle-manager.
                                    Allows for usage of lifecycle-manager without having to worry about modules requiring specific RBAC permissions.
                                    WARNING: DO NOT USE ON PRODUCTIVE CLUSTERS! (default true)
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

* [kyma alpha](kyma_alpha.md)	 - Experimental commands

