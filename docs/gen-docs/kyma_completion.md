---
title: kyma completion
---

Generates bash or zsh completion scripts.

## Synopsis

Use this command to display the shell completion code used for interactive command completion. 
To configure your shell to load completions, add `. <(kyma completion bash)` to your bash profile or `. <(kyma completion zsh)` to your zsh profile.


```bash
kyma completion bash|zsh [flags]
```

## Options inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems.
  -h, --help                Displays help for the command.
      --kubeconfig string   Specifies the path to the kubeconfig file. By default, Kyma CLI uses the KUBECONFIG environment variable or "/$HOME/.kube/config" if the variable is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.

