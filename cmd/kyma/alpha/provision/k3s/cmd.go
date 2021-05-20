package k3s

import (
	"fmt"
	"net"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/k3s"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
		Short:   "Provisions a Kubernetes cluster based on k3s.",
		Long:    `Use this command to provision a k3s-based Kubernetes cluster for Kyma installation.`,
		RunE:    func(_ *cobra.Command, _ []string) error { return c.Run() },
		Aliases: []string{"k"},
	}

	//cmd.Flags().StringVar(&o.EnableRegistry, "enable-registry", "", "Enables registry for the created k8s cluster.")
	cmd.Flags().StringVar(&o.Name, "name", "kyma", `Name of the Kyma cluster`)
	cmd.Flags().IntVar(&o.Workers, "workers", 1, "Number of worker nodes (k3s agents)")
	cmd.Flags().StringSliceVarP(&o.ServerArgs, "server-arg", "s", []string{}, "One or more arguments passed to the Kubernetes API server (e.g. --server-arg='--alsologtostderr')")
	cmd.Flags().StringSliceVarP(&o.AgentArgs, "agent-arg", "a", []string{}, "One or more arguments passed to the k3s agent command on agent nodes (e.g. --agent-arg='--alsologtostderr')")
	cmd.Flags().DurationVar(&o.Timeout, "timeout", 5*time.Minute, `Maximum time for the provisioning. If you want no timeout, enter "0".`)
	cmd.Flags().StringSliceVarP(&o.K3dArgs, "k3d-arg", "", []string{}, "One or more arguments passed to the k3d provisioning command (e.g. --k3d-arg='--no-rollback')")
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
		s.Failure()
		return err
	}

	exists, err := k3s.ClusterExists(c.opts.Verbose, c.opts.Name)
	if err != nil {
		s.Failure()
		return err
	}

	if exists {
		if err := c.deleteExistingK3sCluster(); err != nil {
			s.Failure()
			return err
		}
	} else if err := c.portAllocated(80, 443); err != nil {
		s.Failure()
		return errors.Wrap(err, "Port 80 or 443 cannot be allocated")
	}

	s.Successf("K3s status verified")
	return nil
}

//Check whether a k3s cluster already exists and ensure that all required ports are available
func (c *command) deleteExistingK3sCluster() error {
	var answer bool
	if !c.opts.NonInteractive {
		answer = c.CurrentStep.PromptYesNo("Do you want to remove the existing k3s cluster? ")
		if !answer {
			return fmt.Errorf("User decided not to remove the existing k3s cluster")
		}
	}
	if c.opts.NonInteractive || answer {
		err := k3s.DeleteCluster(c.opts.Verbose, c.opts.Timeout, c.opts.Name)
		if err != nil {
			return err
		}
		c.CurrentStep.Successf("Existing k3s cluster deleted")
	}

	return nil
}

//Check if a port is allocated
func (c *command) portAllocated(ports ...int) error {
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

//Create a k3s cluster
func (c *command) createK3sCluster() error {
	s := c.NewStep("Create K3s instance")
	s.Status("Start K3s cluster")
	err := k3s.StartCluster(c.Verbose, c.opts.Timeout, c.opts.Name, c.opts.Workers, c.opts.ServerArgs, c.opts.AgentArgs, c.opts.K3dArgs)
	if err != nil {
		s.Failuref("Could not start k3s cluster")
		return err
	}
	s.Successf("K3s cluster is created")

	return nil
}

func (c *command) createK3sClusterInfo() error {
	s := c.NewStep("Prepare Kyma installer configuration")
	s.Status("Adding configuration")

	// K8s client needs to be created here because before the kubeconfig is not ready to use
	var err error
	c.K8s, err = kube.NewFromConfig("", c.KubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	clusterInfo := clusterinfo.New(c.K8s.Static())

	if err := clusterInfo.Write(clusterinfo.ClusterProviderK3s, true); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Configuration created")
	return nil
}
