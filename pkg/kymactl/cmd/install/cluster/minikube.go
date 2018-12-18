package cluster

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kyma-incubator/kymactl/internal"
	"github.com/spf13/cobra"
)

const (
	kubernetesVersion string = "1.10.0"
	bootstrapper      string = "localkube"
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
	Domain              string
	VMDriver            string
	Silent              bool
	DiskSize            string
	Memory              string
	CPU                 string
	HypervVirtualSwitch string
}

//NewMinikubeOptions creates options with default values
func NewMinikubeOptions() *MinikubeOptions {
	return &MinikubeOptions{}
}

//NewMinikubeCmd creates a new minikube command
func NewMinikubeCmd(o *MinikubeOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "minikube",
		Short: "Prepares a minikube cluster",
		Long: `Prepares a minikube cluster
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
	cmd.Flags().BoolVarP(&o.Silent, "silent", "s", false, "No interaction")
	return cmd
}

//Run runs the command
func (o *MinikubeOptions) Run() error {
	fmt.Printf("Installing minikube cluster using domain '%s' and vm-driver '%s'\n", o.Domain, o.VMDriver)
	fmt.Println()

	spinner := internal.NewSpinner("Checking requirements", "Requirements are fine")
	if !driverSupported(o.VMDriver) {
		return fmt.Errorf("Specified VMDriver '%s' is not supported by minikube", o.VMDriver)
	}
	if o.VMDriver == vmDriverHyperv && o.HypervVirtualSwitch == "" {
		return fmt.Errorf("Specified VMDriver '%s' requires option --hypervVirtualSwitch to be provided", vmDriverHyperv)
	}
	err := internal.CheckMinikubeVersion()
	if err != nil {
		return err
	}

	err = internal.CheckKubectlVersion()
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	err = checkIfMinikubeIsInitialized(o)
	if err != nil {
		return err
	}

	spinner = internal.NewSpinner("Initializing minikube config", "Minikube config initialized")
	err = initializeMinikubeConfig()
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Waiting for minikube to be up and running", "Minukube up and running")
	err = startMinikube(o)
	if err != nil {
		return err
	}

	err = waitForMinikubeToBeUp()
	if err != nil {
		return err
	}

	err = createClusterRoleBinding()
	if err != nil {
		return err
	}

	err = internal.WaitForPod("kube-system", "k8s-app", "kube-dns")
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	fmt.Println("Adding hostnames, please enter your password if requested")
	err = addDevDomainsToEtcHosts(o)
	if err != nil {
		return err
	}
	fmt.Println("Hostnames added to " + internal.HOSTS_FILE)

	spinner = internal.NewSpinner("Adjusting minikube cluster", "Minikube cluster adjusted")
	err = increaseFsInotifyMaxUserInstances(o)
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	err = printSummary()
	if err != nil {
		return err
	}

	return nil
}

func checkIfMinikubeIsInitialized(o *MinikubeOptions) error {
	statusText, err := internal.RunMinikubeCmdE([]string{"status", "-b=" + bootstrapper, "--format", "'{{.MinikubeStatus}}'"})
	if err != nil {
		return err
	}

	if statusText != "" {
		if !o.Silent {
			fmt.Println("=====")
			fmt.Printf("Minikube is initialized and status is '%s'\n", statusText)
		}
		reader := bufio.NewReader(os.Stdin)
		answer := ""
		if !o.Silent {
			fmt.Printf("Do you want to remove previous minikube cluster [y/N]: ")
			answer, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			fmt.Println("=====")
		}
		if o.Silent || strings.Trim(answer, "\n\t ") == "y" {
			_, err := internal.RunMinikubeCmd([]string{"delete"})
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
	_, err := internal.RunMinikubeCmd([]string{"config", "unset", "ingress"})
	if err != nil {
		return err
	}
	// Enable heapster addon
	_, err = internal.RunMinikubeCmd([]string{"addons", "enable", "heapster"})
	if err != nil {
		return err
	}

	// Disable bootstrapper warning
	_, err = internal.RunMinikubeCmd([]string{"config", "set", "ShowBootstrapperDeprecationNotification", "false"})
	if err != nil {
		return err
	}

	return nil
}

func startMinikube(o *MinikubeOptions) error {
	virtualSwitchArg := ""
	if o.VMDriver == vmDriverHyperv {
		virtualSwitchArg = "--hyperv-virtual-switch=" + o.HypervVirtualSwitch
	}

	startCmd := []string{"start",
		"--memory", o.Memory,
		"--cpus", o.CPU,
		"--extra-config=apiserver.Authorization.Mode=RBAC",
		"--extra-config=apiserver.GenericServerRunOptions.CorsAllowedOriginList='.*'",
		"--extra-config=controller-manager.ClusterSigningCertFile='/var/lib/localkube/certs/ca.crt'",
		"--extra-config=controller-manager.ClusterSigningKeyFile='/var/lib/localkube/certs/ca.key'",
		"--extra-config=apiserver.admission-control='LimitRanger,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota'",
		"--kubernetes-version=v" + kubernetesVersion,
		"--vm-driver=" + o.VMDriver,
		"--disk-size=" + o.DiskSize,
		"--feature-gates='MountPropagation=false'",
		virtualSwitchArg,
		"-b=" + bootstrapper,
	}
	_, err := internal.RunMinikubeCmd(startCmd)
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

func waitForMinikubeToBeUp() error {
	for {
		statusText, err := internal.RunMinikubeCmd([]string{"status", "-b=" + bootstrapper, "--format", "'{{.MinikubeStatus}}'"})
		if err != nil {
			return err
		}

		if statusText == "Running" {
			break
		}
		time.Sleep(sleep)
	}

	for {
		statusText, err := internal.RunMinikubeCmd([]string{"status", "-b=" + bootstrapper, "--format", "'{{.ClusterStatus}}'"})
		if err != nil {
			return err
		}

		if statusText == "Running" {
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

	minikubeIP, err := internal.RunMinikubeCmd([]string{"ip"})
	if err != nil {
		return err
	}

	hostAlias := "127.0.0.1" + hostnames

	if o.VMDriver != vmDriverNone {
		_, err := internal.RunMinikubeCmd([]string{"ssh", "sudo /bin/sh -c 'echo \"" + hostAlias + "\" >> /etc/hosts'"})
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
		_, err := internal.RunMinikubeCmd([]string{"ssh", "--", "'sudo sysctl -w fs.inotify.max_user_instances=524288'"})
		if err != nil {
			return err
		}
	}

	return nil
}

func printSummary() error {
	fmt.Println()
	fmt.Println("Minikube cluster is installed")
	clusterInfo, err := internal.RunMinikubeCmd([]string{"status", "-b=" + bootstrapper})
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
