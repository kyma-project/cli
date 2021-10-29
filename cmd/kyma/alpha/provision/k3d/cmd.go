package k3d

import (
	"fmt"

	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/k3d"

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
		Short:   "Provisions a Kubernetes cluster based on k3d v5.",
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

	k3dClient := k3d.NewClient(k3d.NewCmdRunner(), k3d.NewPathLooker(), c.opts.Name, c.opts.Verbose, c.opts.Timeout, true)
	registryName := fmt.Sprintf(k3d.V5DefaultRegistryNamePattern, c.opts.Name)

	if err = c.verifyK3dStatus(k3dClient, registryName); err != nil {
		return err
	}
	if registryURL, err = c.createK3dRegistry(k3dClient, registryName); err != nil {
		return err
	}
	if err = c.createK3dCluster(k3dClient, registryURL); err != nil {
		return err
	}
	return nil
}

//Verifies if k3d is properly installed and pre-conditions are fulfilled
func (c *command) verifyK3dStatus(k3dClient k3d.Client, registryName string) error {
	s := c.NewStep("Verifying k3d status")
	if err := k3dClient.VerifyStatus(); err != nil {
		s.Failure()
		return err
	}

	s.LogInfo("Checking if port flags are valid")
	ports, err := extractPortsFromFlag(c.opts.PortMapping)
	if err != nil {
		return errors.Wrapf(err, "Could not extract host ports from %s", c.opts.PortMapping)
	}

	s.LogInfo("Checking if k3d registry of previous kyma installation exists")
	registryExists, err := k3dClient.RegistryExists(registryName)
	if err != nil {
		s.Failure()
		return err
	}

	s.LogInfo("Checking if k3d cluster of previous kyma installation exists")
	clusterExists, err := k3dClient.ClusterExists()
	if err != nil {
		s.Failure()
		return err
	}

	if clusterExists {
		if !c.PromptUserToDeleteExistingCluster() {
			s.Failure()
			return fmt.Errorf("User decided not to remove the existing k3d cluster")
		}

		if registryExists {
			if err := k3dClient.DeleteRegistry(registryName); err != nil {
				s.Failure()
				return err
			}
			s.LogInfo("Deleted k3d registry of previous kyma installation")
		}

		if err := k3dClient.DeleteCluster(); err != nil {
			s.Failure()
			return err
		}
		s.LogInfo("Deleted k3d cluster of previous kyma installation")

	} else {
		if err := allocatePorts(ports...); err != nil {
			s.Failure()
			return errors.Wrap(err, "Port cannot be allocated")
		}
		if registryExists {
			// only registry exists
			if err := k3dClient.DeleteRegistry(registryName); err != nil {
				s.Failure()
				return err
			}
			s.LogInfo("Deleted k3d registry of previous kyma installation")
		}
	}

	s.Successf("k3d status verified")
	return nil
}

func (c *command) PromptUserToDeleteExistingCluster() bool {
	var answer bool
	if !c.opts.NonInteractive {
		answer = c.CurrentStep.PromptYesNo("Do you want to remove the existing k3d cluster? ")
	}
	return c.opts.NonInteractive || answer
}

func (c *command) createK3dRegistry(k3dClient k3d.Client, registryName string) (string, error) {
	s := c.NewStep(fmt.Sprintf("Create k3d registry '%s'", registryName))

	registryURL, err := k3dClient.CreateRegistry(registryName)
	if err != nil {
		s.Failuref("Could not create k3d registry")
		return "", err
	}
	s.Successf("Created k3d registry '%s'", registryName)
	return registryURL, nil
}

//Create a k3d cluster
func (c *command) createK3dCluster(k3dClient k3d.Client, registryURL string) error {
	s := c.NewStep(fmt.Sprintf("Create K3d cluster '%s'", c.opts.Name))

	c.opts.UseRegistry = append(c.opts.UseRegistry, registryURL)

	settings := k3d.CreateClusterSettings{
		Args:              parseK3dArgs(c.opts.K3dArgs),
		KubernetesVersion: c.opts.KubernetesVersion,
		PortMapping:       c.opts.PortMapping,
		Workers:           c.opts.Workers,
		V5Settings: k3d.V5CreateClusterSettings{
			K3sArgs:     c.opts.K3sArgs,
			UseRegistry: c.opts.UseRegistry,
		},
	}

	err := k3dClient.CreateCluster(settings)
	if err != nil {
		s.Failuref("Could not create k3d cluster '%s'", c.opts.Name)
		return err
	}
	s.Successf("Created k3d cluster '%s'", c.opts.Name)

	return nil
}

func extractPortsFromFlag(portFlag []string) ([]int, error) {
	var ports []int
	for _, rawPort := range portFlag {
		portStr := rawPort[:strings.IndexByte(rawPort, ':')]
		portInt, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		ports = append(ports, portInt)
	}
	return ports, nil
}

//Check if a port is allocated
func allocatePorts(ports ...int) error {
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

func parseK3dArgs(args []string) []string {
	var res []string
	for _, arg := range args {
		res = append(res, strings.Split(arg, " ")...)
	}
	return res
}
