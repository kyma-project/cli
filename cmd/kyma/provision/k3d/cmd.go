package k3d

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/k3d"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/spf13/cobra"
)

const (
	sleep = 10 * time.Second
)

// k3d create --publish 80:80 --publish 443:443 --enable-registry --registry-volume local_registry --registry-name registry.localhost --server-arg --no-deploy --server-arg traefik -n kyma -t 60
var (
	ErrK3dRunning = errors.New("k3d already running")
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
	// k3d create --publish 80:80 --publish 443:443 --enable-registry --registry-volume local_registry --registry-name registry.localhost --server-arg --no-deploy --server-arg traefik -n kyma -t 60
	cmd := &cobra.Command{
		Use:     "k3d",
		Short:   "Provisions k8s cluster based on k3d.",
		Long:    `Use this command to provision a k3d cluster for Kyma installation.`,
		RunE:    func(_ *cobra.Command, _ []string) error { return c.Run() },
		Aliases: []string{"m"},
	}

	cmd.Flags().StringVar(&o.PublishHTTP, "publishHTTP", "80:80", "Expose ports.")
	cmd.Flags().StringVar(&o.PublishHTTPS, "publishHTTPS", "443:443", "Expose ports.")
	//cmd.Flags().StringVar(&o.EnableRegistry, "enable-registry", "", "Enables registry for the created k8s cluster.")
	cmd.Flags().StringVar(&o.NoDeploy, "no-deploy", "", "No deploy arg.")
	cmd.Flags().StringVar(&o.Name, "name", "kyma", "Name of the Kyma cluster.")
	cmd.Flags().StringVar(&o.RegistryName, "registry-name", "registry.localhost", "Registry name.")
	cmd.Flags().StringVar(&o.RegistryVolume, "registry-volume", "local_registry", "Registry volume.")
	cmd.Flags().StringVar(&o.ServerArg, "server-arg", "traefik", "Server arg.")
	cmd.Flags().DurationVar(&o.Timeout, "timeout", 5*time.Minute, `Maximum time during which the provisioning takes place, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	return cmd
}

//Run runs the command
func (c *command) Run() error {
	s := c.NewStep("Checking K3d status")
	err := c.checkIfK3dIsInitialized(s)
	switch err {
	case ErrK3dRunning, nil:
		break
	default:
		s.Failure()
		return err
	}
	s.Successf("K3d status verified")

	s = c.NewStep("Create K3d instance")
	s.Status("Start K3d")
	err = c.startK3d()
	if err != nil {
		s.Failure()
		return err
	}

	s.Status("Wait for K3d to be up and running")
	err = c.waitForK3dToBeReady(s)
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("K3d cluster is created")

	// K8s client needs to be created here because before the kubeconfig is not ready to use
	if c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath); err != nil {
		log.Println("err: ", err.Error())
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}
	s.Successf("K3d up and running")

	s = c.NewStep("Creating cluster info ConfigMap")
	err = c.createClusterInfoConfigMap()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("ConfigMap created")

	s = c.NewStep("Patch core DNS configuration")
	err = c.PatchCoreDNSConfig()
	if err != nil {
		s.Failure()
		return err
	}
	s.Successf("Configuration for core-dns is patched")
	err = c.printSummary()
	if err != nil {
		return err
	}

	return nil
}

func (c *command) checkIfK3dIsInitialized(s step.Step) error {
	statusText, _ := k3d.RunCmd(c.opts.Timeout, "list")
	if strings.Contains(strings.TrimSpace(statusText), "running") {
		var answer bool
		if !c.opts.NonInteractive {
			answer = s.PromptYesNo("Do you want to remove the existing K3d cluster? ")
		}
		if c.opts.NonInteractive || answer {
			statusText, err := k3d.RunCmd(c.opts.Timeout, "delete", "-a")
			if err != nil {
				return err
			}
			log.Println("Delete the k3d cluster: ", statusText)
		} else {
			return ErrK3dRunning
		}
	}
	return nil
}

func (c *command) startK3d() error {
	startCmd := []string{"create",
		"--publish", c.opts.PublishHTTP,
		"--publish", c.opts.PublishHTTPS,
		"--enable-registry",
		"--registry-volume", c.opts.RegistryVolume,
		"--registry-name", c.opts.RegistryName,
		"--server-arg",
		"--no-deploy",
		"--server-arg", c.opts.ServerArg,
		"--name", c.opts.Name,
	}

	_, err := k3d.RunCmd(c.opts.Timeout, startCmd...)
	if err != nil {
		return err
	}
	return nil
}

func (c *command) setKubeconfigForK3dCluster(step step.Step) error {
	for {
		statusText, err := k3d.RunCmd(c.opts.Timeout, "get-kubeconfig", "-n", c.opts.Name)
		log.Println(statusText)
		if err != nil && !strings.Contains(err.Error(), "Couldnt copy kubeconfig.yaml") {
			return err
		}
		os.Setenv("KUBECONFIG", statusText)
		//step.Status(statusText)
		time.Sleep(sleep)
	}
	return nil
}

func (c *command) waitForK3dToBeReady(step step.Step) error {
	// TODO refactor
	kubeConfigFile := ""
	for {
		statusText, err := k3d.RunCmd(c.opts.Timeout, "get-kubeconfig", "-n", c.opts.Name)
		if err != nil && !strings.Contains(err.Error(), "not ready") {
			return err
		}
		if strings.Contains(strings.TrimSpace(statusText), "kyma/kubeconfig.yaml") {
			kubeConfigFile = strings.Trim(statusText, "\n")
			os.Setenv("KUBECONFIG", kubeConfigFile)
			break
		}
		time.Sleep(sleep)
	}
	for {
		dat, err := ioutil.ReadFile(kubeConfigFile)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				log.Println("err: ", err.Error())
				continue
			}
			log.Println("err: ", err.Error())
			return err
		}
		if strings.Contains(string(dat), "apiVersion") {
			home := os.Getenv("HOME")
			// TODO Append to the list of config and change kubectl config use-context
			copyKubeConfigToDotKubeConfig := exec.Command("cp", fmt.Sprintf("%s/.config/k3d/kyma/kubeconfig.yaml", home), fmt.Sprintf("%s/.kube/config", home))
			output, err := copyKubeConfigToDotKubeConfig.Output()
			log.Println("output: ", string(output))
			if err != nil {
				log.Println("error while copying: ", err)
				return err
			}
			break
		}
		time.Sleep(sleep)
	}

	return nil
}

func (c *command) PatchCoreDNSConfig() error {
	dockerRegIP, err := cli.RunCmd("docker", "inspect", "-f", "'{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'", "/k3d-registry")
	if err != nil {
		return errors.Wrapf(err, "failed to fetch the docker IP for k3d master container")
	}
	dockerRegIP = strings.Trim(dockerRegIP, "\n")
	newCoreDNSConfig := fmt.Sprintf("{\"data\": { \"Corefile\": \"registry.localhost:53 {\\n    hosts {\\n      %s registry.localhost\\n    }\\n}\\n.:53 {\\n    errors\\n    health\\n    rewrite name regex (.*)\\\\.local\\\\.kyma\\\\.pro istio-ingressgateway.istio-system.svc.cluster.local\\n    ready\\n    kubernetes cluster.local in-addr.arpa ip6.arpa {\\n      pods insecure\\n      upstream\\n      fallthrough in-addr.arpa ip6.arpa\\n    }\\n    hosts /etc/coredns/NodeHosts {\\n      reload 1s\\n      fallthrough\\n    }\\n    prometheus :9153\\n    forward . /etc/resolv.conf\\n    cache 30\\n    loop\\n    reload\\n    loadbalance\\n}\"}}", dockerRegIP)

	_, err = c.K8s.Static().CoreV1().ConfigMaps("kube-system").Patch("coredns", types.StrategicMergePatchType, []byte(newCoreDNSConfig))
	if err != nil {
		return errors.Wrapf(err, "failed to patch core DNS configuration")
	}
	return nil
}

func (c *command) printSummary() error {
	//TODO: add total time taken to install
	fmt.Println()
	fmt.Println("K3d cluster installed")
	clusterInfo, err := k3d.RunCmd(c.opts.Timeout, "list")
	if err != nil {
		fmt.Printf("Cannot show cluster-info because of '%s", err)
	} else {
		fmt.Println(clusterInfo)
	}

	fmt.Println("Happy K3d-ing! :)")
	return nil
}

func (c *command) createClusterInfoConfigMap() error {
	cm, err := c.K8s.Static().CoreV1().ConfigMaps("kube-system").Get("kyma-cluster-info", metav1.GetOptions{})
	if err == nil && cm != nil {
		return nil
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	_, err = c.K8s.Static().CoreV1().ConfigMaps("kube-system").Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kyma-cluster-info",
			Labels: map[string]string{"app": "kyma"},
		},
		Data: map[string]string{
			"provider": "k3d",
			"isLocal":  "true",
			"localIP":  "127.0.0.1",
		},
	})

	return err
}
