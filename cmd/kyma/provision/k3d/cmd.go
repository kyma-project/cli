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
		Short:   "Provisions a Kubernetes cluster based on k3d v4.",
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
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.20.11", "Kubernetes version of the cluster")
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

	k3dClient := k3d.NewClient(k3d.NewCmdRunner(), k3d.NewPathLooker(), c.opts.Name, c.opts.Verbose, c.opts.Timeout)

	if err = c.verifyK3dStatus(k3dClient); err != nil {
		return err
	}

	if err = c.createK3dCluster(k3dClient); err != nil {
		return err
	}
	return nil
}

//Verifies if k3d is properly installed and pre-conditions are fulfilled
func (c *command) verifyK3dStatus(k3dClient k3d.Client) error {
	s := c.NewStep("Verifying k3d status")
	if err := k3dClient.VerifyStatus(false); err != nil {
		s.Failure()
		return err
	}

	s.LogInfo("Checking if port flags are valid")
	ports, err := extractPortsFromFlag(c.opts.PortMapping)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Could not extract host ports from %s", c.opts.PortMapping))
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

		if err := k3dClient.DeleteCluster(); err != nil {
			s.Failure()
			return err
		}
		s.LogInfo("Deleted k3d cluster of previous kyma installation")
	} else if err := allocatePorts(ports...); err != nil {
		s.Failure()
		return errors.Wrap(err, "Port cannot be allocated")
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

//Create a k3d cluster
func (c *command) createK3dCluster(k3dClient k3d.Client) error {
	s := c.NewStep(fmt.Sprintf("Create K3d cluster '%s'", c.opts.Name))

	settings := k3d.CreateClusterSettings{
		Args:              parseK3dArgs(c.opts.K3dArgs),
		KubernetesVersion: c.opts.KubernetesVersion,
		PortMapping:       c.opts.PortMapping,
		Workers:           c.opts.Workers,
		V4Settings: k3d.V4CreateClusterSettings{
			ServerArgs: c.opts.ServerArgs,
			AgentArgs:  c.opts.AgentArgs,
		},
	}

	err := k3dClient.CreateCluster(settings, false)
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
