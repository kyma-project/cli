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
	cmd.Flags().StringSliceVarP(&o.K3sArgs, "k3s-arg", "s", []string{}, "One or more arguments passed from k3d to the k3s command (format: ARG@NODEFILTER[;@NODEFILTER])")
	cmd.Flags().DurationVar(&o.Timeout, "timeout", 5*time.Minute, `Maximum time for the provisioning. If you want no timeout, enter "0".`)
	cmd.Flags().StringSliceVarP(&o.K3dArgs, "k3d-arg", "", []string{}, "One or more arguments passed to the k3d provisioning command (e.g. --k3d-arg='--no-rollback')")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.20.11", "Kubernetes version of the cluster")
	cmd.Flags().StringSliceVar(&o.UseRegistry, "registry-use", []string{}, "Connect to one or more k3d-managed registries. Kyma automatically creates a registry for serverless images.")
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

	var err error
	var registryURL string

	if err = c.verifyK3dStatus(); err != nil {
		return err
	}
	if registryURL, err = c.createK3dRegistry(); err != nil {
		return err
	}
	if err = c.createK3dCluster(registryURL); err != nil {
		return err
	}
	if err = c.createK3dClusterInfo(); err != nil {
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

	registryExists, err := k3d.RegistryExists(c.opts.Verbose, c.opts.Name)
	if err != nil {
		return err
	}

	clusterExists, err := k3d.ClusterExists(c.opts.Verbose, c.opts.Name)
	if err != nil {
		s.Failure()
		return err
	}
	ports, err := extractPortsFromFlag(c.opts.PortMapping)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Could not extract host ports from %s", c.opts.PortMapping))
	}

	if clusterExists {
		if err := c.deleteExistingK3dCluster(registryExists); err != nil {
			s.Failure()
			return err
		}
	} else {
		if err := c.allocatePorts(ports...); err != nil {
			s.Failure()
			return errors.Wrap(err, "Port cannot be allocated")
		}
		if registryExists {
			// only registry exists
			if err := k3d.DeleteRegistry(c.opts.Verbose, c.opts.Timeout, c.opts.Name); err != nil {
				s.Failure()
				return err
			}
		}
	}

	s.Successf("K3d status verified")
	return nil
}

//Deletes the existing k3d cluster and deletes the k3d registry (if one exists)
func (c *command) deleteExistingK3dCluster(registryExists bool) error {
	var answer bool
	if !c.opts.NonInteractive {
		answer = c.CurrentStep.PromptYesNo("Do you want to remove the existing k3d cluster? ")
		if !answer {
			return fmt.Errorf("User decided not to remove the existing k3d cluster")
		}
	}
	if c.opts.NonInteractive || answer {
		// if the default registry exists, delete it also
		if registryExists {
			if err := k3d.DeleteRegistry(c.opts.Verbose, c.opts.Timeout, c.opts.Name); err != nil {
				return err
			}
		}

		if err := k3d.DeleteCluster(c.opts.Verbose, c.opts.Timeout, c.opts.Name); err != nil {
			return err
		}
		c.CurrentStep.Successf("Existing k3d cluster deleted")
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

func (c *command) createK3dRegistry() (string, error) {
	s := c.NewStep("Create K3d registry")

	registryURL, err := k3d.CreateRegistry(c.Verbose, c.opts.Timeout, c.opts.Name)
	if err != nil {
		s.Failuref("Could not create k3d registry")
		return "", err
	}
	s.Successf("K3d registry is created")
	return registryURL, nil
}

//Create a k3d cluster
func (c *command) createK3dCluster(registryURL string) error {
	s := c.NewStep("Create K3d instance")
	s.Status("Start K3d cluster")

	k3dSettings := k3d.Settings{
		ClusterName: c.opts.Name,
		Args:        parseK3dargs(c.opts.K3dArgs),
		Version:     c.opts.KubernetesVersion,
		PortMapping: c.opts.PortMapping,
	}
	c.opts.UseRegistry = append(c.opts.UseRegistry, registryURL)

	err := k3d.StartCluster(c.Verbose, c.opts.Timeout, c.opts.Workers, c.opts.K3sArgs, c.opts.UseRegistry, k3dSettings)
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
