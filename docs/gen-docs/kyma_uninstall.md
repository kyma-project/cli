## kyma uninstall

Uninstalls Kyma from a running Kubernetes cluster.

### Synopsis

Use this command to uninstall Kyma from a running Kubernetes cluster.

Make sure that your kubeconfig file is already pointing to the target cluster.<br>

This command:
- Removes your cluster administrator account.
- Removes Tiller.
- Removes the Kyma Installer.

### Usage


```
kyma uninstall [flags]
```

### Options

```
  -h, --help               Displays help for the command.
      --timeout duration   Time-out after which Kyma CLI stops watching the the process of unstalling Kyma. (default 30m0s)
```

### Options inherited from parent commands

```
      --kubeconfig string   Specifies the path to the kubeconfig file. (default "/$HOME/.kube/config")
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

### SEE ALSO

* [kyma](kyma.md)	 - Controls a Kyma cluster.

