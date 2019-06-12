package minikube

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/minikube"
	"github.com/kyma-project/cli/internal/step"
	"github.com/kyma-project/cli/pkg/kyma/core"

	"github.com/spf13/cobra"
)

const (
	kubernetesVersion  string = "1.12.5"
	bootstrapper       string = "kubeadm"
	vmDriverHyperkit   string = "hyperkit"
	vmDriverHyperv     string = "hyperv"
	vmDriverNone       string = "none"
	vmDriverVirtualBox string = "virtualbox"
	sleep                     = 10 * time.Second
)

var (
	domains = []string{
		"apiserver",
		"console",
		"catalog",
		"instances",
		"brokers",
		"dex",
		"docs",
		"lambdas-ui",
		"console-backend",
		"minio",
		"jaeger",
		"grafana",
		"configurations-generator",
		"gateway",
		"connector-service",
		"log-ui",
		"loki",
		"kiali",
	}

	drivers = []string{
		"virtualbox",
		"vmwarefusion",
		"kvm",
		"xhyve",
		vmDriverHyperv,
		vmDriverHyperkit,
		"kvm2",
		"none",
	}
	ErrMinikubeRunning = errors.New("Minikube already running")
)

type command struct {
	opts *options
	core.Command
}

//NewCmd creates a new minikube command
func NewCmd(o *options) *cobra.Command {

	c := command{
		Command: core.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:     "minikube",
		Short:   "Provisions Minikube",
		Long:    `Provisions Minikube for Kyma installation`,
		RunE:    func(_ *cobra.Command, _ []string) error { return c.Run() },
		Aliases: []string{"m"},
	}

	cmd.Flags().StringVarP(&o.Domain, "domain", "d", "kyma.local", "domain to use")
	cmd.Flags().StringVar(&o.VMDriver, "vm-driver", defaultVMDriver, "VMDriver to use, possible values are: "+strings.Join(drivers, ","))
	cmd.Flags().StringVar(&o.HypervVirtualSwitch, "hypervVirtualSwitch", "", "Name of the hyperv switch, required if --vm-driver=hyperv")
	cmd.Flags().StringVar(&o.DiskSize, "disk-size", "30g", "Disk size to use")
	cmd.Flags().StringVar(&o.Memory, "memory", "8192", "Memory to use")
	cmd.Flags().StringVar(&o.CPUS, "cpus", "4", "CPUs to use")
	return cmd
}

//Run runs the command
func (c *command) Run() error {
	s := c.NewStep("Checking requirements")
	err := c.checkRequirements(s)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements verified")

	s.LogInfof("Preparing Minikube using domain '%s' and vm-driver '%s'", c.opts.Domain, c.opts.VMDriver)

	s = c.NewStep("Check Minikube status")
	err = c.checkIfMinikubeIsInitialized(s)
	switch err {
	case ErrMinikubeRunning, nil:
		break
	default:
		s.Failure()
		return err
	}
	s.Successf("Minikube status verified")

	s = c.NewStep(fmt.Sprintf("Initializing Minikube config"))
	err = c.initializeMinikubeConfig()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minikube config initialized")

	s = c.NewStep(fmt.Sprintf("Create Minikube instance"))
	s.Status("Start Minikube")
	err = c.startMinikube()
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Wait for Minikube to be up and running")
	err = c.waitForMinikubeToBeUp(s)
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Create default cluster role")
	err = c.createClusterRoleBinding()
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Wait for kube-dns to be up and running")
	err = c.Kubectl().WaitForPodReady("kube-system", "k8s-app", "kube-dns")
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minikube up and running")

	err = c.addDevDomainsToEtcHosts(s)
	if err != nil {
		return err
	}

	s = c.NewStep(fmt.Sprintf("Adjusting Minikube cluster"))
	s.Status("Increase fs.inotify.max_user_instances")
	err = c.increaseFsInotifyMaxUserInstances()
	if err != nil {
		s.Failure()
		return err
	}
	s.Status("Enable metrics server")
	err = c.enableMetricsServer()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Adjustments finished")

	err = c.printSummary()
	if err != nil {
		return err
	}

	return nil
}

func (c *command) checkRequirements(s step.Step) error {
	if !driverSupported(c.opts.VMDriver) {
		s.Failure()
		return fmt.Errorf("Specified VMDriver '%s' is not supported by Minikube", c.opts.VMDriver)
	}
	if c.opts.VMDriver == vmDriverHyperv && c.opts.HypervVirtualSwitch == "" {
		s.Failure()
		return fmt.Errorf("Specified VMDriver '%s' requires the --hypervVirtualSwitch option", vmDriverHyperv)
	}

	versionWarning, err := minikube.CheckVersion(c.opts.Verbose)
	if err != nil {
		s.Failure()
		return err
	}
	if versionWarning != "" {
		s.LogError(versionWarning)
	}

	versionWarning, err = c.Kubectl().CheckVersion()
	if err != nil {
		s.Failure()
		return err
	}
	if versionWarning != "" {
		s.LogError(versionWarning)
	}
	return nil
}

func (c *command) checkIfMinikubeIsInitialized(s step.Step) error {
	statusText, _ := minikube.RunCmd(c.opts.Verbose, "status", "-b", bootstrapper, "--format", "{{.Host}}")

	if strings.TrimSpace(statusText) != "" {
		var answer string
		var err error
		if !c.opts.NonInteractive {
			answer, err = s.Prompt("Do you want to remove the existing Minikube cluster? [y/N]: ")
			if err != nil {
				return err
			}
		}
		if c.opts.NonInteractive || strings.TrimSpace(answer) == "y" {
			_, err := minikube.RunCmd(c.opts.Verbose, "delete")
			if err != nil {
				return err
			}
		} else {
			return ErrMinikubeRunning
		}
	}
	return nil
}

func (c *command) initializeMinikubeConfig() error {
	// Disable default nginx ingress controller
	_, err := minikube.RunCmd(c.opts.Verbose, "config", "unset", "ingress")
	if err != nil {
		return err
	}
	return nil
}

func (c *command) startMinikube() error {
	startCmd := []string{"start",
		"--memory", c.opts.Memory,
		"--cpus", c.opts.CPUS,
		"--extra-config=apiserver.authorization-mode=RBAC",
		"--extra-config=apiserver.cors-allowed-origins='http://*'",
		"--extra-config=apiserver.enable-admission-plugins=DefaultStorageClass,LimitRanger,MutatingAdmissionWebhook,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,ValidatingAdmissionWebhook",
		"--kubernetes-version=v" + kubernetesVersion,
		"--vm-driver", c.opts.VMDriver,
		"--disk-size", c.opts.DiskSize,
		"-b", bootstrapper,
	}

	if c.opts.VMDriver == vmDriverHyperv {
		startCmd = append(startCmd, "--hyperv-virtual-switch="+c.opts.HypervVirtualSwitch)

	}
	_, err := minikube.RunCmd(c.opts.Verbose, startCmd...)
	if err != nil {
		return err
	}
	return nil
}

// fixes https://github.com/kyma-project/kyma/issues/1986
func (c *command) createClusterRoleBinding() error {
	check, err := c.Kubectl().IsClusterResourceDeployed("clusterrolebinding", "app", "kyma")
	if err != nil {
		return err
	}
	if !check {
		_, err := c.Kubectl().RunCmd("create", "clusterrolebinding", "default-sa-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default")
		if err != nil {
			return err
		}
		_, err = c.Kubectl().RunCmd("label", "clusterrolebinding", "default-sa-cluster-admin", "app=kyma")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *command) waitForMinikubeToBeUp(step step.Step) error {
	for {
		statusText, err := minikube.RunCmd(c.opts.Verbose, "status", "-b="+bootstrapper, "--format", "'{{.Host}}'")
		if err != nil {
			return err
		}
		step.Status(statusText)

		if strings.TrimSpace(statusText) == "Running" {
			break
		}
		time.Sleep(sleep)
	}

	for {
		statusText, err := minikube.RunCmd(c.opts.Verbose, "status", "-b="+bootstrapper, "--format", "'{{.Kubelet}}'")
		if err != nil {
			return err
		}
		step.Status(statusText)

		if strings.TrimSpace(statusText) == "Running" {
			break
		}
		time.Sleep(sleep)
	}

	return nil
}

func (c *command) addDevDomainsToEtcHosts(s step.Step) error {
	hostnames := ""
	for _, v := range domains {
		hostnames = hostnames + " " + v + "." + c.opts.Domain
	}

	minikubeIP, err := minikube.RunCmd(c.opts.Verbose, "ip")
	if err != nil {
		return err
	}

	hostAlias := "127.0.0.1" + hostnames

	if c.opts.VMDriver != vmDriverNone {
		_, err := minikube.RunCmd(c.opts.Verbose, "ssh", "sudo /bin/sh -c 'echo \""+hostAlias+"\" >> /etc/hosts'")
		if err != nil {
			return err
		}
	}

	hostAlias = strings.Trim(minikubeIP, "\n") + hostnames

	return addDevDomainsToEtcHostsOSSpecific(c.opts.Domain, s, hostAlias)
}

// Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
func (c *command) increaseFsInotifyMaxUserInstances() error {
	if c.opts.VMDriver != vmDriverNone {
		_, err := minikube.RunCmd(c.opts.Verbose, "ssh", "--", "sudo sysctl -w fs.inotify.max_user_instances=524288")
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *command) enableMetricsServer() error {
	_, err := minikube.RunCmd(c.opts.Verbose, "addons", "enable", "metrics-server")
	if err != nil {
		return err
	}
	return nil
}

func (c *command) printSummary() error {
	fmt.Println()
	fmt.Println("Minikube cluster is installed")
	clusterInfo, err := minikube.RunCmd(c.opts.Verbose, "status", "-b="+bootstrapper)
	if err != nil {
		fmt.Printf("Cannot show cluster-info because of '%s", err)
	} else {
		fmt.Println(clusterInfo)
	}

	fmt.Println("Happy Minikube-ing! :)")
	return nil
}

func driverSupported(driver string) bool {
	for _, element := range drivers {
		if element == driver {
			return true
		}
	}
	return false
}
