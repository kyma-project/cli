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
	minikubeVersion   string = "0.28.2"
	kubectlVersion    string = "1.10.0"
	kubernetesVersion string = "1.10.0"
	bootstrapper      string = "localkube"
	vmDriverHyperkit  string = "hyperkit"
	vmDriverNone      string = "none"
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

//MinikubeOptions defines available options for the command
type MinikubeOptions struct {
	Domain   string
	VMDriver string
	Silent   bool
	DiskSize string
	Memory   string
	CPU      string
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
	cmd.Flags().StringVar(&o.VMDriver, "vm-driver", vmDriverHyperkit, "VMDriver to use")
	cmd.Flags().StringVar(&o.DiskSize, "disk-size", "20g", "Disk size to use")
	cmd.Flags().StringVar(&o.Memory, "memory", "8192", "Memory to use")
	cmd.Flags().StringVar(&o.CPU, "cpu", "4", "CPUs to use")
	cmd.Flags().BoolVarP(&o.Silent, "silent", "s", false, "No interaction")
	return cmd
}

//Run runs the command
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

	fmt.Println("\nHappy Minikube-ing!")
	clusterInfoCmd := []string{"status", "-b=" + bootstrapper}
	clusterInfo, err := internal.RunMinikubeCmd(clusterInfoCmd)
	if err != nil {
		fmt.Printf("Cannot show cluster-info because of '%s", err)
	} else {
		fmt.Println(clusterInfo)
	}

	return nil
}

func checkMinikubeVersion() error {
	versionCmd := []string{"version"}
	versionText, err := internal.RunMinikubeCmd(versionCmd)
	if err != nil {
		return err
	}

	exp, _ := regexp.Compile("minikube version: v((\\d+.\\d+.\\d+))")
	version := exp.FindStringSubmatch(versionText)

	if version[1] != minikubeVersion {
		return fmt.Errorf("Currently minikube in version '%s' is required", minikubeVersion)
	}
	return nil

}

func checkKubectlVersion() error {
	versionCmd := []string{"version", "--client", "--short"}
	versionText, err := internal.RunKubectlCmd(versionCmd)
	if err != nil {
		return err
	}

	exp, _ := regexp.Compile("Client Version: v((\\d+).(\\d+).(\\d+))")
	kubctlIsVersion := exp.FindStringSubmatch(versionText)

	exp, _ = regexp.Compile("((\\d+).(\\d+).(\\d+))")
	kubctlMustVersion := exp.FindStringSubmatch(kubectlVersion)

	majorIsVersion, _ := strconv.Atoi(kubctlIsVersion[2])
	majorMustVersion, _ := strconv.Atoi(kubctlMustVersion[2])
	minorIsVersion, _ := strconv.Atoi(kubctlIsVersion[3])
	minorMustVersion, _ := strconv.Atoi(kubctlMustVersion[3])

	if minorIsVersion-minorMustVersion < -1 || minorIsVersion-minorMustVersion > 1 {
		fmt.Printf("Your kubectl version is '%s'. Supported versions of kubectl are from '%d.%d.*' to '%d.%d.*'", kubctlIsVersion[1], majorMustVersion, minorMustVersion-1, majorMustVersion, minorMustVersion+1)
	}
	if majorIsVersion != majorMustVersion {
		return fmt.Errorf("Your kubectl version is '%s'. Supported versions of kubectl are from '%d.%d.*' to '%d.%d.*'", kubctlIsVersion[1], majorMustVersion, minorMustVersion-1, majorMustVersion, minorMustVersion+1)
	}
	return nil
}

func checkIfMinikubeIsInitialized(o *MinikubeOptions) error {
	statusCmd := []string{"status", "-b=" + bootstrapper, "--format", "'{{.MinikubeStatus}}'"}
	statusText, err := internal.RunMinikubeCmdE(statusCmd)
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
		if o.Silent || answer == "y\n" {
			deleteCmd := []string{"delete"}
			_, err := internal.RunMinikubeCmd(deleteCmd)
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
	ingressConfigCmd := []string{"config", "unset", "ingress"}
	_, err := internal.RunMinikubeCmd(ingressConfigCmd)
	if err != nil {
		return err
	}
	// Enable heapster addon
	ingressHeapsterCmd := []string{"addons", "enable", "heapster"}
	_, err = internal.RunMinikubeCmd(ingressHeapsterCmd)
	if err != nil {
		return err
	}

	// Disable bootstrapper warning
	bootstrapperConfigCmd := []string{"config", "set", "ShowBootstrapperDeprecationNotification", "false"}
	_, err = internal.RunMinikubeCmd(bootstrapperConfigCmd)
	if err != nil {
		return err
	}

	return nil
}

func startMinikube(o *MinikubeOptions) error {
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
		"-b=" + bootstrapper,
	}
	_, err := internal.RunMinikubeCmd(startCmd)
	if err != nil {
		return err
	}
	return nil
}

// fixes https://github.com/kyma-project/kyma/issues/1986
func applyHotfix() error {
	hotfixCmd := []string{"create", "clusterrolebinding", "kube-system-cluster-admin", "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default"}
	_, err := internal.RunKubectlCmd(hotfixCmd)
	if err != nil {
		fmt.Printf("\nTried to fix minikube setup for https://github.com/kyma-project/kyma/issues/1986 but failed")
		fmt.Println(err)
	}
	return nil
}

func waitForMinikubeToBeUp() error {
	for {
		statusCmd := []string{"status", "-b=" + bootstrapper, "--format", "'{{.MinikubeStatus}}'"}
		statusText, err := internal.RunMinikubeCmd(statusCmd)
		if err != nil {
			return err
		}

		if statusText == "Running" {
			break
		}
		time.Sleep(sleep)
	}

	for {
		statusCmd := []string{"status", "-b=" + bootstrapper, "--format", "'{{.MinikubeStatus}}'"}
		statusText, err := internal.RunMinikubeCmd(statusCmd)
		if err != nil {
			return err
		}

		if statusText == "Running" {
			break
		}
		time.Sleep(sleep)
	}

	err := applyHotfix()
	if err != nil {
		return err
	}
	err = internal.WaitForPod("kube-system", "k8s-app", "kube-dns")
	if err != nil {
		return err
	}

	return nil
}

func addDevDomainsToEtcHosts(o *MinikubeOptions) error {
	hostnames := ""
	for _, v := range domains {
		hostnames = hostnames + " " + v + "." + o.Domain
	}

	cmd := []string{"ip"}
	minikubeIP, err := internal.RunMinikubeCmd(cmd)
	if err != nil {
		return err
	}

	if o.VMDriver != vmDriverNone {
		cmd := []string{"ssh", "'echo \"127.0.0.1" + hostnames + "\" | sudo tee -a /etc/hosts'"}
		_, err := internal.RunMinikubeCmd(cmd)
		if err != nil {
			return err
		}
	}

	hostAlias := minikubeIP + hostnames

	cmd = []string{hostAlias, "|", "sudo", "tee", "-a", "/etc/hosts", ">", "/dev/null"}
	_, err = internal.RunCmd("echo", cmd)
	if err != nil {
		return err
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
		cmd := []string{"ssh", "--", "'sudo sysctl -w fs.inotify.max_user_instances=524288'"}
		_, err := internal.RunMinikubeCmd(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}
