package k3d

import (
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/k3d"
	"github.com/kyma-project/cli/internal/kube"

	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new k3d command
func NewCmd(o *Options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:     "k3d",
		Short:   "Provisions a Kubernetes cluster based on k3d.",
		Long:    `Use this command to provision a k3d-based Kubernetes cluster for Kyma installation.`,
		RunE:    func(_ *cobra.Command, _ []string) error { return c.Run() },
		Aliases: []string{"k"},
	}

	cmd.Flags().StringVar(&o.Name, "name", "kyma", `Name of the Kyma cluster`)
	cmd.Flags().IntVar(&o.Workers, "workers", 1, "Number of worker nodes (k3d agents)")
	cmd.Flags().StringSliceVarP(&o.ServerArgs, "server-arg", "s", []string{}, "One or more arguments passed to the Kubernetes API server (e.g. --server-arg='--alsologtostderr')")
	cmd.Flags().StringSliceVarP(&o.AgentArgs, "agent-arg", "a", []string{}, "One or more arguments passed to the k3d agent command on agent nodes (e.g. --agent-arg='--alsologtostderr')")
	cmd.Flags().DurationVar(&o.Timeout, "timeout", 5*time.Minute, `Maximum time for the provisioning. If you want no timeout, enter "0".`)
	cmd.Flags().StringSliceVarP(&o.K3dArgs, "k3d-arg", "", []string{}, "One or more arguments passed to the k3d provisioning command (e.g. --k3d-arg='--no-rollback')")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.20.7", "Kubernetes version of the cluster")
	cmd.Flags().StringSliceVarP(&o.PortMapping, "port", "p", []string{"80:80@loadbalancer", "443:443@loadbalancer"}, "Map ports 80 and 443 of K3D loadbalancer (e.g. -p 80:80@loadbalancer -p 443:443@loadbalancer)")
	return cmd
}

//Run runs the command
func (c *command) Run() error {
	if c.opts.CI {
		c.Factory.NonInteractive = true
	}
	if c.opts.Verbose {
		c.Factory.UseLogger = true
	}

	if err := c.verifyK3dStatus(); err != nil {
		return err
	}
	if err := c.createK3dCluster(); err != nil {
		return err
	}
	if err := c.createK3dClusterInfo(); err != nil {
		return err
	}
	return nil
}

func extractPortsFromFlag(portFlag []string) ([]int, error) {
	ports := []int{}
	for _, rawport := range portFlag {
		portStr := rawport[:strings.IndexByte(rawport, ':')]
		portInt, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		ports = append(ports, portInt)
	}
	return ports, nil
}

//Ensure k3d is installed and pre-conditions are fulfilled
func (c *command) verifyK3dStatus() error {
	s := c.NewStep("Checking k3d status")
	if err := k3d.Initialize(c.Verbose); err != nil {
		s.Failure()
		return err
	}

	exists, err := k3d.ClusterExists(c.opts.Verbose, c.opts.Name)
	if err != nil {
		s.Failure()
		return err
	}
	ports, err := extractPortsFromFlag(c.opts.PortMapping)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Could not extract host ports from %s", c.opts.PortMapping))
	}

	if exists {
		if err := c.deleteExistingK3dCluster(); err != nil {
			s.Failure()
			return err
		}
	} else if err := c.allocatePorts(ports...); err != nil {
		s.Failure()
		return errors.Wrap(err, "Port cannot be allocated")
	}

	s.Successf("K3d status verified")
	return nil
}

//Check whether a k3d cluster already exists and ensure that all required ports are available
func (c *command) deleteExistingK3dCluster() error {
	var answer bool
	if !c.opts.NonInteractive {
		answer = c.CurrentStep.PromptYesNo("Do you want to remove the existing k3d cluster? ")
		if !answer {
			return fmt.Errorf("User decided not to remove the existing k3d cluster")
		}
	}
	if c.opts.NonInteractive || answer {
		err := k3d.DeleteCluster(c.opts.Verbose, c.opts.Timeout, c.opts.Name)
		if err != nil {
			return err
		}
		c.CurrentStep.Successf("Existing k3s cluster deleted")
	}

	return nil
}

//Check if a port is allocated
func (c *command) allocatePorts(ports ...int) error {
	for _, port := range ports {
		con, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}
		if con != nil {
			con.Close()
		}
	}
	return nil
}

func parseK3dargs(args []string) []string {
	var res []string
	for _, arg := range args {
		res = append(res, strings.Split(arg, " ")...)
	}
	return res
}

//Create a k3d cluster
func (c *command) createK3dCluster() error {
	s := c.NewStep("Create K3d instance")
	s.Status("Start K3d cluster")

	k3dSettings := k3d.Settings{
		ClusterName: c.opts.Name,
		Args:        parseK3dargs(c.opts.K3dArgs),
		Version:     c.opts.KubernetesVersion,
		PortMapping: c.opts.PortMapping,
	}
	err := k3d.StartCluster(c.Verbose, c.opts.Timeout, c.opts.Workers, c.opts.ServerArgs, c.opts.AgentArgs, k3dSettings)
	if err != nil {
		s.Failuref("Could not start k3d cluster")
		return err
	}
	s.Successf("K3d cluster is created")

	return nil
}

func (c *command) createK3dClusterInfo() error {
	s := c.NewStep("Prepare Kyma installer configuration")
	s.Status("Adding configuration")

	// K8s client needs to be created here because before the kubeconfig is not ready to use
	var err error
	c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	clusterInfo := clusterinfo.New(c.K8s.Static())

	if err := clusterInfo.Write(clusterinfo.ClusterProviderK3d, true); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Configuration created")
	return nil
}
