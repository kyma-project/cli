package minikube

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/kubectl"
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

//MinikubeOptions defines available options for the command
type MinikubeOptions struct {
	*core.Options
	Domain              string
	VMDriver            string
	DiskSize            string
	Memory              string
	CPU                 string
	HypervVirtualSwitch string
}

//NewOptions creates options with default values
func NewOptions(o *core.Options) *MinikubeOptions {
	return &MinikubeOptions{Options: o}
}

//NewCmd creates a new minikube command
func NewCmd(o *MinikubeOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "minikube",
		Short:   "Provisions Minikube",
		Long:    `Provisions Minikube for Kyma installation`,
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Aliases: []string{"m"},
	}

	cmd.Flags().StringVarP(&o.Domain, "domain", "d", "kyma.local", "domain to use")
	cmd.Flags().StringVar(&o.VMDriver, "vm-driver", defaultVMDriver, "VMDriver to use, possible values are: "+strings.Join(drivers, ","))
	cmd.Flags().StringVar(&o.HypervVirtualSwitch, "hypervVirtualSwitch", "", "Name of the hyperv switch, required if --vm-driver=hyperv")
	cmd.Flags().StringVar(&o.DiskSize, "disk-size", "30g", "Disk size to use")
	cmd.Flags().StringVar(&o.Memory, "memory", "8192", "Memory to use")
	cmd.Flags().StringVar(&o.CPU, "cpu", "4", "CPUs to use")
	return cmd
}

//Run runs the command
func (o *MinikubeOptions) Run() error {
	s := o.NewStep("Checking requirements")
	err := checkRequirements(o, s)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements verified")

	s.LogInfof("Preparing Minikube using domain '%s' and vm-driver '%s'", o.Domain, o.VMDriver)

	s = o.NewStep("Check Minikube status")
	err = checkIfMinikubeIsInitialized(o, s)
	switch err {
	case ErrMinikubeRunning, nil:
		break
	default:
		s.Failure()
		return err
	}
	s.Successf("Minikube status verified")

	s = o.NewStep(fmt.Sprintf("Initializing Minikube config"))
	err = initializeMinikubeConfig(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minikube config initialized")

	s = o.NewStep(fmt.Sprintf("Create Minikube instance"))
	s.Status("Start Minikube")
	err = startMinikube(o)
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Wait for Minikube to be up and running")
	err = waitForMinikubeToBeUp(o, s)
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Create default cluster role")
	err = createClusterRoleBinding(o)
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Wait for kube-dns to be up and running")

	err = kubectl.WaitForPodReady("kube-system", "k8s-app", "kube-dns", o.Verbose)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minikube up and running")

	err = addDevDomainsToEtcHosts(o, s)
	if err != nil {
		return err
	}

	s = o.NewStep(fmt.Sprintf("Adjusting Minikube cluster"))
	s.Status("Increase fs.inotify.max_user_instances")
	err = increaseFsInotifyMaxUserInstances(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Status("Enable metrics server")
	err = enableMetricsServer(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Adjustments finished")

	err = printSummary(o)
	if err != nil {
		return err
	}

	return nil
}

func checkRequirements(o *MinikubeOptions, s step.Step) error {
	if !driverSupported(o.VMDriver) {
		s.Failure()
		return fmt.Errorf("Specified VMDriver '%s' is not supported by Minikube", o.VMDriver)
	}
	if o.VMDriver == vmDriverHyperv && o.HypervVirtualSwitch == "" {
		s.Failure()
		return fmt.Errorf("Specified VMDriver '%s' requires the --hypervVirtualSwitch option", vmDriverHyperv)
	}

	versionWarning, err := minikube.CheckVersion(o.Verbose)
	if err != nil {
		s.Failure()
		return err
	}
	if versionWarning != "" {
		s.LogError(versionWarning)
	}

	versionWarning, err = kubectl.CheckVersion(o.Verbose)
	if err != nil {
		s.Failure()
		return err
	}
	if versionWarning != "" {
		s.LogError(versionWarning)
	}
	return nil
}

func checkIfMinikubeIsInitialized(o *MinikubeOptions, s step.Step) error {
	statusText, _ := minikube.RunCmd(o.Verbose, "status", "-b", bootstrapper, "--format", "{{.Host}}")

	if strings.TrimSpace(statusText) != "" {
		var answer string
		var err error
		if !o.NonInteractive {
			answer, err = s.Prompt("Do you want to remove the existing Minikube cluster? [y/N]: ")
			if err != nil {
				return err
			}
		}
		if o.NonInteractive || strings.TrimSpace(answer) == "y" {
			_, err := minikube.RunCmd(o.Verbose, "delete")
			if err != nil {
				return err
			}
		} else {
			return ErrMinikubeRunning
		}
	}
	return nil
}

func initializeMinikubeConfig(o *MinikubeOptions) error {
	// Disable default nginx ingress controller
	_, err := minikube.RunCmd(o.Verbose, "config", "unset", "ingress")
	if err != nil {
		return err
	}
	return nil
}

func startMinikube(o *MinikubeOptions) error {
	startCmd := []string{"start",
		"--memory", o.Memory,
		"--cpus", o.CPU,
		"--extra-config=apiserver.authorization-mode=RBAC",
		"--extra-config=apiserver.cors-allowed-origins='http://*'",
		"--extra-config=apiserver.enable-admission-plugins=DefaultStorageClass,LimitRanger,MutatingAdmissionWebhook,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,ValidatingAdmissionWebhook",
		"--kubernetes-version=v" + kubernetesVersion,
		"--vm-driver", o.VMDriver,
		"--disk-size", o.DiskSize,
		"-b", bootstrapper,
	}

	if o.VMDriver == vmDriverHyperv {
		startCmd = append(startCmd, "--hyperv-virtual-switch="+o.HypervVirtualSwitch)

	}
	_, err := minikube.RunCmd(o.Verbose, startCmd...)
	if err != nil {
		return err
	}
	return nil
}

// fixes https://github.com/kyma-project/kyma/issues/1986
func createClusterRoleBinding(o *MinikubeOptions) error {
	check, err := kubectl.IsClusterResourceDeployed("clusterrolebinding", "app", "kyma", o.Verbose)
	if err != nil {
		return err
	}
	if !check {
		_, err := kubectl.RunCmd(o.Verbose, "create", "clusterrolebinding", "default-sa-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default")
		if err != nil {
			return err
		}
		_, err = kubectl.RunCmd(o.Verbose, "label", "clusterrolebinding", "default-sa-cluster-admin", "app=kyma")
		if err != nil {
			return err
		}
	}
	return nil
}

func waitForMinikubeToBeUp(o *MinikubeOptions, step step.Step) error {
	for {
		statusText, err := minikube.RunCmd(o.Verbose, "status", "-b="+bootstrapper, "--format", "'{{.Host}}'")
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
		statusText, err := minikube.RunCmd(o.Verbose, "status", "-b="+bootstrapper, "--format", "'{{.Kubelet}}'")
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

func addDevDomainsToEtcHosts(o *MinikubeOptions, s step.Step) error {
	hostnames := ""
	for _, v := range domains {
		hostnames = hostnames + " " + v + "." + o.Domain
	}

	minikubeIP, err := minikube.RunCmd(o.Verbose, "ip")
	if err != nil {
		return err
	}

	hostAlias := "127.0.0.1" + hostnames

	if o.VMDriver != vmDriverNone {
		_, err := minikube.RunCmd(o.Verbose, "ssh", "sudo /bin/sh -c 'echo \""+hostAlias+"\" >> /etc/hosts'")
		if err != nil {
			return err
		}
	}

	hostAlias = strings.Trim(minikubeIP, "\n") + hostnames

	return addDevDomainsToEtcHostsOSSpecific(o, s, hostAlias)
}

// Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
func increaseFsInotifyMaxUserInstances(o *MinikubeOptions) error {
	if o.VMDriver != vmDriverNone {
		_, err := minikube.RunCmd(o.Verbose, "ssh", "--", "sudo sysctl -w fs.inotify.max_user_instances=524288")
		if err != nil {
			return err
		}
	}

	return nil
}

func enableMetricsServer(o *MinikubeOptions) error {
	_, err := minikube.RunCmd(o.Verbose, "addons", "enable", "metrics-server")
	if err != nil {
		return err
	}
	return nil
}

func printSummary(o *MinikubeOptions) error {
	fmt.Println()
	fmt.Println("Minikube cluster is installed")
	clusterInfo, err := minikube.RunCmd(o.Verbose, "status", "-b="+bootstrapper)
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
