package provision

import (
	"bufio"
	"fmt"
	"github.com/kyma-incubator/kyma-cli/internal/minikube"
	"github.com/kyma-incubator/kyma-cli/internal/step"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kyma-incubator/kyma-cli/internal"
	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"
	"github.com/spf13/cobra"
)

const (
	kubernetesVersion string = "1.11.5"
	bootstrapper      string = "kubeadm"
	vmDriverHyperkit  string = "hyperkit"
	vmDriverHyperv    string = "hyperv"
	vmDriverNone      string = "none"
	sleep                    = 10 * time.Second
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
		"ui-api",
		"minio",
		"jaeger",
		"grafana",
		"configurations-generator",
		"gateway",
		"connector-service",
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

//NewMinikubeOptions creates options with default values
func NewMinikubeOptions(o *core.Options) *MinikubeOptions {
	return &MinikubeOptions{Options: o}
}

//NewMinikubeCmd creates a new minikube command
func NewMinikubeCmd(o *MinikubeOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minikube",
		Short: "Provisions minikube",
		Long: `Provisions minikube for Kyma installation
`,
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Aliases: []string{"m"},
	}

	cmd.Flags().StringVarP(&o.Domain, "domain", "d", "kyma.local", "domain to use")
	cmd.Flags().StringVar(&o.VMDriver, "vm-driver", vmDriverHyperkit, "VMDriver to use, possible values are: "+strings.Join(drivers, ","))
	cmd.Flags().StringVar(&o.HypervVirtualSwitch, "hypervVirtualSwitch", "", "Name of the hyperv switch to use, required if --vm-driver=hyperv")
	cmd.Flags().StringVar(&o.DiskSize, "disk-size", "20g", "Disk size to use")
	cmd.Flags().StringVar(&o.Memory, "memory", "8192", "Memory to use")
	cmd.Flags().StringVar(&o.CPU, "cpu", "4", "CPUs to use")
	return cmd
}

//Run runs the command
func (o *MinikubeOptions) Run() error {
	fmt.Printf("Preparing minikube using domain '%s' and vm-driver '%s'\n", o.Domain, o.VMDriver)
	fmt.Println()

	s := o.NewStep(fmt.Sprintf("Checking requirements"))
	if !driverSupported(o.VMDriver) {
		s.Failure()
		return fmt.Errorf("Specified VMDriver '%s' is not supported by minikube", o.VMDriver)
	}
	if o.VMDriver == vmDriverHyperv && o.HypervVirtualSwitch == "" {
		s.Failure()
		return fmt.Errorf("Specified VMDriver '%s' requires option --hypervVirtualSwitch to be provided", vmDriverHyperv)
	}

	isSupportedMinikube, err := minikube.CheckVersion()
	if err != nil {
		s.Failure()
		return err
	}
	if !isSupportedMinikube {
		s.LogError("You are using unsupported minikube version. This may not work.")
	}

	err = internal.CheckKubectlVersion()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements are fine")

	err = checkIfMinikubeIsInitialized(o)
	if err != nil {
		return err
	}

	s = o.NewStep(fmt.Sprintf("Initializing minikube config"))
	err = initializeMinikubeConfig()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minikube config initialized")

	s = o.NewStep(fmt.Sprintf("Create minikube instance"))
	s.Status("Start minikube")
	err = startMinikube(o)
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Await minikube to be up and running")
	err = waitForMinikubeToBeUp(s)
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Create default cluster role")
	err = createClusterRoleBinding()
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Await kube-dns to be up and running")
	err = internal.WaitForPod("kube-system", "k8s-app", "kube-dns")
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minukube up and running")

	fmt.Println("Adding hostnames, please enter your password if requested")
	err = addDevDomainsToEtcHosts(o)
	if err != nil {
		return err
	}
	fmt.Println("Hostnames added to " + internal.HOSTS_FILE)

	s = o.NewStep(fmt.Sprintf("Adjusting minikube cluster"))
	err = increaseFsInotifyMaxUserInstances(o)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minikube cluster adjusted")

	err = printSummary()
	if err != nil {
		return err
	}

	return nil
}

func checkIfMinikubeIsInitialized(o *MinikubeOptions) error {
	statusText, err := minikube.RunCmd("status", "-b", bootstrapper)
	if err != nil {
		return err
	}

	if strings.TrimSpace(statusText) != "" {
		reader := bufio.NewReader(os.Stdin)
		answer := ""
		if !o.NonInteractive {
			fmt.Printf("Do you want to remove previous minikube cluster [y/N]: ")
			answer, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
		}
		if o.NonInteractive || strings.TrimSpace(answer) == "y" {
			_, err := minikube.RunCmd("delete")
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("minikube installation cancelled")
		}
	}
	return nil
}

func initializeMinikubeConfig() error {
	// Disable default nginx ingress controller
	_, err := minikube.RunCmd("config", "unset", "ingress")
	if err != nil {
		return err
	}
	// Enable heapster addon
	_, err = minikube.RunCmd("addons", "enable", "heapster")
	if err != nil {
		return err
	}

	// Disable bootstrapper warning
	_, err = minikube.RunCmd("config", "set", "ShowBootstrapperDeprecationNotification", "false")
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
		startCmd = append(startCmd, "--hyperv-virtual-switch=" + o.HypervVirtualSwitch)

	}
	_, err := minikube.RunCmd(startCmd...)
	if err != nil {
		return err
	}
	return nil
}

// fixes https://github.com/kyma-project/kyma/issues/1986
func createClusterRoleBinding() error {
	check, err := internal.IsClusterResourceDeployed("clusterrolebinding", "app", "kyma")
	if err != nil {
		return err
	}
	if !check {
		_, err := internal.RunKubectlCmd([]string{"create", "clusterrolebinding", "default-sa-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default"})
		if err != nil {
			return err
		}
		_, err = internal.RunKubectlCmd([]string{"label", "clusterrolebinding", "default-sa-cluster-admin", "app=kyma"})
		if err != nil {
			return err
		}
	}
	return nil
}

func waitForMinikubeToBeUp(step step.Step) error {
	for {
		statusText, err := minikube.RunCmd("status", "-b=" + bootstrapper, "--format", "'{{.Host}}'")
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
		statusText, err := minikube.RunCmd("status", "-b=" + bootstrapper, "--format", "'{{.Kubelet}}'")
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

func addDevDomainsToEtcHosts(o *MinikubeOptions) error {
	hostnames := ""
	for _, v := range domains {
		hostnames = hostnames + " " + v + "." + o.Domain
	}

	minikubeIP, err := minikube.RunCmd("ip")
	if err != nil {
		return err
	}

	hostAlias := "127.0.0.1" + hostnames

	if o.VMDriver != vmDriverNone {
		_, err := minikube.RunCmd("ssh", "sudo /bin/sh -c 'echo \"" + hostAlias + "\" >> /etc/hosts'")
		if err != nil {
			return err
		}
	}

	hostAlias = strings.Trim(minikubeIP, "\n") + hostnames

	if runtime.GOOS == "windows" {
		fmt.Println()
		fmt.Println("=====")
		fmt.Println("Please add these lines to your " + internal.HOSTS_FILE + " file:")
		fmt.Println(hostAlias)
		fmt.Println("=====")
	} else {
		_, err := internal.RunCmd("sudo", []string{"/bin/sh", "-c", "sed -i '' \"/" + o.Domain + "/d\" " + internal.HOSTS_FILE})
		if err != nil {
			return err
		}

		_, err = internal.RunCmd("sudo", []string{"/bin/sh", "-c", "echo '" + hostAlias + "' >> " + internal.HOSTS_FILE})
		if err != nil {
			return err
		}
	}

	/* does not work because of permission denied
	f, err := os.OpenFile(internal.HOSTS_FILE, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = f.WriteString(hostAlias)
	if err != nil {
		return err
	}*/

	return nil
}

// Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
func increaseFsInotifyMaxUserInstances(o *MinikubeOptions) error {
	if o.VMDriver != vmDriverNone {
		_, err := minikube.RunCmd("ssh", "--", "'sudo sysctl -w fs.inotify.max_user_instances=524288'")
		if err != nil {
			return err
		}
	}

	return nil
}

func printSummary() error {
	fmt.Println()
	fmt.Println("Minikube cluster is installed")
	clusterInfo, err := minikube.RunCmd("status", "-b=" + bootstrapper)
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
