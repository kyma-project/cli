package minikube

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/minikube"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/spf13/cobra"
)

const (
	bootstrapper       string = "kubeadm"
	vmDriverHyperkit   string = "hyperkit"
	vmDriverHyperv     string = "hyperv"
	vmDriverDocker     string = "docker"
	vmDriverNone       string = "none"
	vmDriverVirtualBox string = "virtualbox"
	sleep                     = 10 * time.Second
)

var (
	drivers = []string{
		"vmwarefusion",
		"kvm",
		"xhyve",
		vmDriverHyperv,
		vmDriverHyperkit,
		vmDriverVirtualBox,
		"kvm2",
		vmDriverDocker,
		"none",
	}
	ErrMinikubeRunning = errors.New("Minikube already running")
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new minikube command
func NewCmd(o *Options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:     "minikube",
		Short:   "Provisions Minikube.",
		Long:    `Use this command to provision a Minikube cluster for Kyma installation. It requires to have Minikube installed upfront, see also https://github.com/kubernetes/minikube`,
		RunE:    func(_ *cobra.Command, _ []string) error { return c.Run() },
		Aliases: []string{"m"},
	}

	cmd.Flags().StringVar(&o.VMDriver, "vm-driver", defaultVMDriver, "Specifies the VM driver. Possible values: "+strings.Join(drivers, ","))
	cmd.Flags().StringVar(&o.HypervVirtualSwitch, "hyperv-virtual-switch", "", "Specifies the Hyper-V switch version if you choose Hyper-V as the driver.")
	cmd.Flags().StringSliceVar(&o.DockerPorts, "docker-ports", []string{}, "List of ports that should be exposed if you choose Docker as the driver.")
	cmd.Flags().StringVar(&o.DiskSize, "disk-size", "30g", "Specifies the disk size used for installation.")
	cmd.Flags().StringVar(&o.Memory, "memory", "8192", "Specifies RAM reserved for installation.")
	cmd.Flags().StringVar(&o.CPUS, "cpus", "4", "Specifies the number of CPUs used for installation.")
	cmd.Flags().StringVar(&o.Profile, "profile", "", "Specifies the Minikube profile.")
	cmd.Flags().DurationVar(&o.Timeout, "timeout", 5*time.Minute, `Maximum time during which the provisioning takes place, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	cmd.Flags().BoolVar(&o.UseVPNKitSock, "use-hyperkit-vpnkit-sock", false, `Uses vpnkit sock provided by Docker. This is useful when DNS Port (53) is being used by some other program like dns-proxy (eg. provided by Cisco Umbrella. This flag works only on Mac OS).`)
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.16.15", "Kubernetes version of the cluster.")
	return cmd
}

//Run runs the command
func (c *command) Run() error {
	s := c.NewStep("Checking requirements")
	if err := c.checkRequirements(s); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Requirements verified")

	s.LogInfof("Preparing Minikube using vm-driver '%s'", c.opts.VMDriver)

	s = c.NewStep("Checking Minikube status")
	err := c.checkIfMinikubeIsInitialized(s)
	switch err {
	case ErrMinikubeRunning, nil:
		break
	default:
		s.Failure()
		return err
	}
	s.Successf("Minikube status verified")

	s = c.NewStep("Initializing Minikube config")
	err = c.initializeMinikubeConfig()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minikube config initialized")

	s = c.NewStep("Create Minikube instance")
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

	// K8s client needs to be created here because before the kubeconfig is not ready to use
	if c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	s.Status("Create default cluster role")
	err = c.createClusterRoleBinding()
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Wait for kube-dns to be up and running")
	err = c.K8s.WaitPodStatusByLabel("kube-system", "k8s-app", "kube-dns", corev1.PodRunning)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Minikube up and running")

	s = c.NewStep("Adjusting Minikube cluster")
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

	s = c.NewStep("Creating cluster info ConfigMap")
	err = c.createClusterInfoConfigMap()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("ConfigMap created")

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
		return fmt.Errorf("Specified VMDriver '%s' requires the --hyperv-virtual-switch option", vmDriverHyperv)
	}

	if len(c.opts.DockerPorts) > 0 && c.opts.VMDriver != vmDriverDocker {
		s.Failure()
		return fmt.Errorf("docker-ports flag is applicable only for VMDriver '%s'", vmDriverDocker)
	}

	versionWarning, err := minikube.CheckVersion(c.opts.Verbose, c.opts.Timeout)
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
	statusText, _ := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "status", "-b", bootstrapper, "--format", "{{.Host}}")

	if strings.TrimSpace(statusText) != "" {
		var answer bool
		if !c.opts.NonInteractive {
			answer = s.PromptYesNo("Do you want to remove the existing Minikube cluster? ")
		}
		if c.opts.NonInteractive || answer {
			_, err := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "delete")
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
	_, err := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "config", "unset", "ingress")
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
		"--extra-config=apiserver.service-account-signing-key-file=/var/lib/minikube/certs/sa.key",
		"--extra-config=apiserver.service-account-issuer=kubernetes/serviceaccount",
		"--extra-config=apiserver.service-account-api-audiences=api",
		"--kubernetes-version=v" + c.opts.KubernetesVersion,
		"--vm-driver", c.opts.VMDriver,
		"--disk-size", c.opts.DiskSize,
		"-b", bootstrapper,
	}

	if c.opts.VMDriver == vmDriverHyperv {
		startCmd = append(startCmd, "--hyperv-virtual-switch="+c.opts.HypervVirtualSwitch)
	}

	if c.opts.VMDriver == vmDriverDocker && len(c.opts.DockerPorts) > 0 {
		for _, port := range c.opts.DockerPorts {
			startCmd = append(startCmd, "--ports="+port)
		}
	}

	startCmd, err := osSpecificRun(c, startCmd)
	if err != nil {
		return err
	}
	_, err = minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, startCmd...)
	if err != nil {
		return err
	}
	return nil
}

// fixes https://github.com/kyma-project/kyma/issues/1986
func (c *command) createClusterRoleBinding() error {
	var err error
	bs, err := c.K8s.Static().RbacV1().ClusterRoleBindings().List(context.Background(), metav1.ListOptions{LabelSelector: "app=kyma"})
	if err != nil {
		return err
	}
	if len(bs.Items) == 0 {
		_, err = c.K8s.Static().RbacV1().ClusterRoleBindings().Create(context.Background(), &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "default-sa-cluster-admin",
				Labels: map[string]string{"app": "kyma"},
			},
			RoleRef: rbacv1.RoleRef{
				Name: "cluster-admin",
				Kind: "ClusterRole",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Namespace: "kube-system",
					Name:      "default",
				},
			},
		}, metav1.CreateOptions{})
	}
	return err
}

func (c *command) waitForMinikubeToBeUp(step step.Step) error {
	for {
		statusText, err := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "status", "-b="+bootstrapper, "--format", "'{{.Host}}'")
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
		statusText, err := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "status", "-b="+bootstrapper, "--format", "'{{.Kubelet}}'")
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

// Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
func (c *command) increaseFsInotifyMaxUserInstances() error {
	if c.opts.VMDriver != vmDriverNone {
		_, err := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "ssh", "--", "sudo sysctl -w fs.inotify.max_user_instances=524288")
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *command) enableMetricsServer() error {
	_, err := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "addons", "enable", "metrics-server")
	if err != nil {
		return err
	}
	return nil
}

func (c *command) printSummary() error {
	fmt.Println()
	fmt.Println("Minikube cluster installed")
	clusterInfo, err := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "status", "-b="+bootstrapper)
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

func (c *command) createClusterInfoConfigMap() error {
	cm, err := c.K8s.Static().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kyma-cluster-info", metav1.GetOptions{})
	if err == nil && cm != nil {
		return nil
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	minikubeIP := c.getMinikubeIP()
	_, err = c.K8s.Static().CoreV1().ConfigMaps("kube-system").Create(context.Background(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kyma-cluster-info",
			Labels: map[string]string{"app": "kyma"},
		},
		Data: map[string]string{
			"provider":      "minikube",
			"isLocal":       "true",
			"profile":       c.opts.Profile,
			"localIP":       minikubeIP,
			"localVMDriver": c.opts.VMDriver,
		},
	}, metav1.CreateOptions{})

	return err
}

func (c *command) getMinikubeIP() string {
	minikubeIP, err := minikube.RunCmd(c.opts.Verbose, c.opts.Profile, c.opts.Timeout, "ip")
	if err != nil {
		c.CurrentStep.LogInfo("Unable to perform 'minikube ip' command. IP won't be passed to Kyma")
		return ""
	}
	return strings.TrimSpace(minikubeIP)
}
