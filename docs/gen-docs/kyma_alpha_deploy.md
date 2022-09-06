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
      --dry-run               Renders the Kubernetes manifests without actually applying them.
  -m, --module stringArray    Provide one or more modules to activate after the deployment is finished, for example:
                              	- With short-hand notation: "--module name@namespace"
                              	- With verbose JSON structure "--module '{"name": "componentName","namespace": "componenNamespace","url": "componentUrl","version": "1.2.3"}'
  -f, --modules-file string   Path to file containing a list of modules.
  -s, --source string         Installation source:
                              	- Deploy a specific release  of the lifecycle and module manager: "kyma deploy --source=2.0.0"
                              	- Deploy a specific branch of the lifecycle and module manager: "kyma deploy --source=<my-branch-name>"
                              	- Deploy a commit (8 characters or more) of the lifecycle and module manager: "kyma deploy --source=34edf09a"
                              	- Deploy a pull request, for example "kyma deploy --source=PR-9486"
                              	- Deploy the local sources  of the lifecycle and module manager: "kyma deploy --source=local" (default "main")
  -t, --timeout duration      Maximum time for the deployment. (default 20m0s)
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

* [kyma alpha](kyma_alpha.md)	 - Experimental commands

