---
title: kyma alpha deploy
---

Deploys Kyma on a running Kubernetes cluster.

## Synopsis

Use this command to deploy, upgrade, or adapt Kyma on a running Kubernetes cluster.

```bash
kyma alpha deploy [flags]
```

## Flags

```bash
  -c, --channel string              Select which channel to deploy from. (default "regular")
      --dry-run                     Renders the Kubernetes manifests without actually applying them.
  -k, --kustomization stringArray   Provide one or more kustomizations to deploy. Each occurrence of the flag accepts a URL with an optional reference (commit, branch, or release) in the format URL@ref or a local path to the directory of the kustomization file.
                                    	Defaults to deploying Lifecycle Manager and Module Manager from GitHub main branch.
                                    	Examples:
                                    	- Deploy a specific release of the Lifecycle Manager: "kyma deploy -k https://github.com/kyma-project/lifecycle-manager/config/default@1.2.3"
                                    	- Deploy a local Module Manager: "kyma deploy --kustomization /path/to/repo/module-manager/config/default"
                                    	- Deploy a branch of Lifecycle Manager with a custom URL: "kyma deploy -k https://gitlab.com/forked-from-github/lifecycle-manager/config/default@feature-branch-1"
                                    	- Deploy the main branch of Lifecycle Manager while using local sources of Module Manager: "kyma deploy -k /path/to/repo/module-manager/config/default -k https://github.com/kyma-project/lifecycle-manager/config/default@main"
      --kyma-cr string              Provide a custom Kyma CR file for the deployment.
  -m, --module stringArray          Provide one or more modules to activate after the deployment is finished. Example: "--module name@namespace" (namespace is optional).
  -f, --modules-file string         Path to file containing a list of modules.
  -n, --namespace string            The Namespace to deploy the the Kyma custom resource in. (default "kyma-system")
      --template stringArray        Provide one or more module templates to deploy.
                                    	WARNING: This is a temporary flag for development and will be removed soon.
  -t, --timeout duration            Maximum time for the deployment. (default 20m0s)
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

