## kyma provision minikube

Provisions Minikube.

### Synopsis

Use this command to provision Minikube for Kyma installation.

```
kyma provision minikube [flags]
```

### Options

```
      --cpus string                  Specifies the number of CPUs used for installation. (default "4")
      --disk-size string             Specifies the disk size used for installation. (default "30g")
-h,   --help                         Displays help for the command.
      --hypervVirtualSwitch string   Specifies the Hyper-V switch version if you choose Hyper-V as the driver.
      --memory string                Specifies RAM reserved for installation. (default "8192")
      --profile string               Specifies the minikube profile. (default "minikube")
      --vm-driver string             Specifies the VM driver. Possible values: virtualbox,vmwarefusion,kvm,xhyve,hyperv,hyperkit,kvm2,none (default "hyperkit")
```

### Options inherited from parent commands

```
      --kubeconfig string   Specifies the path to the kubeconfig file. Use the default KUBECONFIG environment variable or /$HOME/.kube/config if KUBECONFIG is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

### SEE ALSO

* [kyma provision](kyma_provision.md)	 - Provisions a cluster for Kyma installation.


