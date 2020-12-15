package k3s

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/k3s"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new k3s command
func NewCmd(o *Options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:     "k3s",
		Short:   "Provisions k8s cluster based on k3s.",
		Long:    `Use this command to provision a k3s cluster for Kyma installation.`,
		RunE:    func(_ *cobra.Command, _ []string) error { return c.Run() },
		Aliases: []string{"k"},
	}

	//cmd.Flags().StringVar(&o.EnableRegistry, "enable-registry", "", "Enables registry for the created k8s cluster.")
	cmd.Flags().StringVar(&o.Name, "name", "kyma", "Name of the Kyma cluster.")
	cmd.Flags().IntVar(&o.Workers, "workers", 1, "Number of worker nodes.")
	cmd.Flags().DurationVar(&o.Timeout, "timeout", 5*time.Minute, `Maximum time in minutes which the provisioning takes place, where "0" means "infinite".`)
	return cmd
}

//Run runs the command
func (c *command) Run() error {
	if err := c.verifyK3sStatus(); err != nil {
		return err
	}
	if err := c.createK3sCluster(); err != nil {
		return err
	}
	if err := c.createK3sClusterInfo(); err != nil {
		return err
	}
	return nil
}

//Ensure k3s is installed and pre-conditions are fulfilled
func (c *command) verifyK3sStatus() error {
	s := c.NewStep("Checking k3s status")
	if err := k3s.Initialize(c.Verbose); err != nil {
		return err
	}

	if c.portAllocated(80) || c.portAllocated(443) {
		s.Failuref("Port 80 or 443 are already in use. Please stop the allocating service and try again.")
	}

	if err := c.checkIfK3sInitialized(s); err != nil {
		s.Failure()
	}
	s.Successf("K3s status verified")

	return nil
}

//Check if a port is allocated
func (c *command) portAllocated(port int) bool {
	con, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if con != nil {
		con.Close()
	}
	return err != nil
}

//Check whether a k3s cluster already exists
func (c *command) checkIfK3sInitialized(s step.Step) error {
	exists, err := k3s.ClusterExists(c.opts.Verbose, c.opts.Name)
	if err != nil {
		return err
	}

	if exists {
		var answer bool
		if !c.opts.NonInteractive {
			answer = s.PromptYesNo("Do you want to remove the existing k3s cluster? ")
		}
		if c.opts.NonInteractive || answer {
			err := k3s.DeleteCluster(c.opts.Verbose, c.opts.Timeout, c.opts.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//Create a k3s cluster
func (c *command) createK3sCluster() error {
	s := c.NewStep("Create K3s instance")
	s.Status("Start K3s cluster")
	err := k3s.StartCluster(c.Verbose, c.opts.Timeout, c.opts.Name)
	if err != nil {
		s.Failuref("Could not start k3s cluster")
		return err
	}
	s.Successf("K3d cluster is created")

	// K8s client needs to be created here because before the kubeconfig is not ready to use
	c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	return nil
}

func (c *command) createK3sClusterInfo() error {
	s := c.NewStep("Prepare Kyma installer configuration")
	s.Status("Adding configuration")
	if err := c.createClusterInfoConfigMap(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Configuration created")
	return nil
}

//Insert Kyma installer configuration as config-map
func (c *command) createClusterInfoConfigMap() error {
	cm, err := c.K8s.Static().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kyma-cluster-info", metav1.GetOptions{})
	if err == nil && cm != nil {
		return nil
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	_, err = c.K8s.Static().CoreV1().ConfigMaps("kube-system").Create(context.Background(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kyma-cluster-info",
			Labels: map[string]string{"app": "kyma"},
		},
		Data: map[string]string{
			"provider": "k3d",
			"isLocal":  "true",
			"localIP":  "127.0.0.1",
		},
	}, metav1.CreateOptions{})

	return err
}
