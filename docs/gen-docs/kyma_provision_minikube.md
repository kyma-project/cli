---
title: kyma provision minikube
---

[DEPRECATED] Provisions Minikube.

## Synopsis

[DEPRECATED: The "provision minikube" command is deprecated. Use the "provision k3d" command instead.]

Use this command to provision a Minikube cluster for Kyma installation. It requires to have Minikube installed upfront, see also https://github.com/kubernetes/minikube

```bash
kyma provision minikube [flags]
```

## Flags

```bash
      --cpus string                    Specifies the number of CPUs used for installation. (default "4")
      --disk-size string               Specifies the disk size used for installation. (default "30g")
      --docker-ports strings           List of ports that should be exposed if you choose Docker as the driver.
      --hyperv-virtual-switch string   Specifies the Hyper-V switch version if you choose Hyper-V as the driver.
  -k, --kube-version string            Kubernetes version of the cluster. (default "1.16.15")
      --memory string                  Specifies RAM reserved for installation. (default "8192")
      --profile string                 Specifies the Minikube profile.
      --timeout duration               Maximum time during which the provisioning takes place, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 5m0s)
      --use-hyperkit-vpnkit-sock       Uses vpnkit sock provided by Docker. This is useful when DNS Port (53) is being used by some other program like dns-proxy (eg. provided by Cisco Umbrella. This flag works only on Mac OS).
      --vm-driver string               Specifies the VM driver. Possible values: vmwarefusion,kvm,xhyve,hyperv,hyperkit,virtualbox,kvm2,docker,none (default "hyperkit")
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Command help
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma provision](#kyma-provision-kyma-provision)	 - Provisions a cluster for Kyma installation.

