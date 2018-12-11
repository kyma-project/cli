package cluster

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"time"

	"strconv"

	"github.com/kyma-incubator/kymactl/internal"
	"github.com/spf13/cobra"
)

const (
	MINIKUBE_VERSION   string = "0.28.2"
	KUBECTL_VERSION    string = "1.10.0"
	KUBERNETES_VERSION string = "1.10.0"
	BOOTSTRAPPER       string = "localkube"
	VM_DRIVER_HYPERKIT string = "hyperkit"
	VM_DRIVER_NONE     string = "none"
)

const (
	sleep = 10 * time.Second
)

var (
	domains = [...]string{
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
)

type MinikubeOptions struct {
	Domain   string
	VMDriver string
	Silent   bool
	DiskSize string
	Memory   string
	Cpu      string
}

func NewMinikubeOptions() *MinikubeOptions {
	return &MinikubeOptions{}
}

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
	cmd.Flags().StringVar(&o.VMDriver, "vm-driver", VM_DRIVER_HYPERKIT, "VMDriver to use")
	cmd.Flags().StringVar(&o.DiskSize, "disk-size", "20g", "Disk size to use")
	cmd.Flags().StringVar(&o.Memory, "memory", "8192", "Memory to use")
	cmd.Flags().StringVar(&o.Cpu, "cpu", "4", "CPUs to use")
	cmd.Flags().BoolVarP(&o.Silent, "silent", "s", false, "No interaction")
	return cmd
}

func (o *MinikubeOptions) Run() error {
	fmt.Printf("Installing minikube cluster using domain '%s' and vm-driver '%s'\n\n", o.Domain, o.VMDriver)

	spinner := internal.NewSpinner("Checking requirements", "Requirements are fine")
	err := checkMinikubeVersion()
	if err != nil {
		return err
	}

	err = checkKubectlVersion()
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
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Adding hostnames to /etc/hosts", "Hostnames added to /etc/hosts")
	err = addDevDomainsToEtcHosts(o)
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	spinner = internal.NewSpinner("Adjusting minikube cluster", "Minikube cluster adjusted")
	err = increaseFsInotifyMaxUserInstances(o)
	if err != nil {
		return err
	}
	internal.StopSpinner(spinner)

	fmt.Println("\nHappy Minikube-ing!\n")
	clusterInfoCmd := []string{"status", "-b=" + BOOTSTRAPPER}
	clusterInfo := internal.RunMinikubeCmd(clusterInfoCmd)

	fmt.Println(clusterInfo)

	return nil
}

func checkMinikubeVersion() error {

	versionCmd := []string{"version"}
	versionText := internal.RunMinikubeCmd(versionCmd)

	exp, _ := regexp.Compile("minikube version: v((\\d+.\\d+.\\d+))")
	version := exp.FindStringSubmatch(versionText)

	if version[1] != MINIKUBE_VERSION {
		return fmt.Errorf("Currently minikube in version '%s' is required", MINIKUBE_VERSION)
	}
	return nil

}

func checkKubectlVersion() error {
	versionCmd := []string{"version", "--client", "--short"}
	versionText := internal.RunKubeCmd(versionCmd)

	exp, _ := regexp.Compile("Client Version: v((\\d+).(\\d+).(\\d+))")
	kubctlIsVersion := exp.FindStringSubmatch(versionText)

	exp, _ = regexp.Compile("((\\d+).(\\d+).(\\d+))")
	kubctlMustVersion := exp.FindStringSubmatch(KUBECTL_VERSION)

	majorIsVersion, _ := strconv.Atoi(kubctlIsVersion[2])
	majorMustVersion, _ := strconv.Atoi(kubctlMustVersion[2])
	minorIsVersion, _ := strconv.Atoi(kubctlIsVersion[3])
	minorMustVersion, _ := strconv.Atoi(kubctlMustVersion[3])

	if minorIsVersion-minorMustVersion < -1 || minorIsVersion-minorMustVersion > 1 {
		return fmt.Errorf("Your kubectl version is '%s'. Supported versions of kubectl are from '%d.%d.*' to '%d.%d.*'", kubctlIsVersion[1], majorMustVersion, minorMustVersion-1, majorMustVersion, minorMustVersion+1)
	}
	if majorIsVersion != majorMustVersion {
		return fmt.Errorf("Your kubectl version is '%s'. Supported versions of kubectl are from '%d.%d.*' to '%d.%d.*'", kubctlIsVersion[1], majorMustVersion, minorMustVersion-1, majorMustVersion, minorMustVersion+1)
	}
	return nil
}

func checkIfMinikubeIsInitialized(o *MinikubeOptions) error {
	statusCmd := []string{"status", "-b=" + BOOTSTRAPPER, "--format", "'{{.MinikubeStatus}}'"}
	statusText := internal.RunMinikubeCmdE(statusCmd)

	if statusText != "" {
		if !o.Silent {
			fmt.Println("=====")
			fmt.Printf("Minikube is initialized and status is '%s'\n", statusText)
		}
		reader := bufio.NewReader(os.Stdin)
		answer := ""
		if !o.Silent {
			fmt.Printf("Do you want to remove previous minikube cluster [y/N]: ")
			answer, _ = reader.ReadString('\n')
			fmt.Println("=====")
		}
		if o.Silent || answer == "y\n" {
			deleteCmd := []string{"delete"}
			internal.RunMinikubeCmd(deleteCmd)
		} else {
			return fmt.Errorf("minikube installation cancelled")
		}
	}
	return nil
}

func initializeMinikubeConfig() error {
	// Disable default nginx ingress controller
	ingressConfigCmd := []string{"config", "unset", "ingress"}
	internal.RunMinikubeCmdE(ingressConfigCmd)
	// Enable heapster addon
	ingressHeapsterCmd := []string{"addons", "enable", "heapster"}
	internal.RunMinikubeCmdE(ingressHeapsterCmd)

	// Disable bootstrapper warning
	bootstrapperConfigCmd := []string{"config", "set", "ShowBootstrapperDeprecationNotification", "false"}
	internal.RunMinikubeCmdE(bootstrapperConfigCmd)

	return nil
}

func startMinikube(o *MinikubeOptions) error {
	startCmd := []string{"start",
		"--memory", o.Memory,
		"--cpus", o.Cpu,
		"--extra-config=apiserver.Authorization.Mode=RBAC",
		"--extra-config=apiserver.GenericServerRunOptions.CorsAllowedOriginList='.*'",
		"--extra-config=controller-manager.ClusterSigningCertFile='/var/lib/localkube/certs/ca.crt'",
		"--extra-config=controller-manager.ClusterSigningKeyFile='/var/lib/localkube/certs/ca.key'",
		"--extra-config=apiserver.admission-control='LimitRanger,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota'",
		"--kubernetes-version=v" + KUBERNETES_VERSION,
		"--vm-driver=" + o.VMDriver,
		"--disk-size=" + o.DiskSize,
		"--feature-gates='MountPropagation=false'",
		"-b=" + BOOTSTRAPPER,
	}
	internal.RunMinikubeCmd(startCmd)
	return nil
}

func waitForMinikubeToBeUp() error {
	for {
		statusCmd := []string{"status", "-b=" + BOOTSTRAPPER, "--format", "'{{.MinikubeStatus}}'"}
		statusText := internal.RunMinikubeCmdE(statusCmd)

		if statusText == "Running" {
			break
		}
		time.Sleep(sleep)
	}

	for {
		statusCmd := []string{"status", "-b=" + BOOTSTRAPPER, "--format", "'{{.MinikubeStatus}}'"}
		statusText := internal.RunMinikubeCmdE(statusCmd)

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

	cmd := []string{"ip"}
	minikubeIP := internal.RunMinikubeCmd(cmd)

	if o.VMDriver != VM_DRIVER_NONE {
		cmd := []string{"ssh", "'echo \"127.0.0.1" + hostnames + "\" | sudo tee -a /etc/hosts'"}
		internal.RunMinikubeCmd(cmd)
	}

	hostAlias := minikubeIP + hostnames

	cmd = []string{hostAlias, "|", "sudo", "tee", "-a", "/etc/hosts", ">", "/dev/null"}
	internal.RunCmd("echo", cmd)

	return nil
}

// Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
func increaseFsInotifyMaxUserInstances(o *MinikubeOptions) error {
	if o.VMDriver != VM_DRIVER_NONE {
		cmd := []string{"ssh", "--", "'sudo sysctl -w fs.inotify.max_user_instances=524288'"}
		internal.RunMinikubeCmd(cmd)
	}

	return nil
}
