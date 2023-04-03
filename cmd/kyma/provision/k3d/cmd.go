package k3d

import (
	"fmt"

	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-project/cli/cmd/kyma/provision"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/k3d"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

// NewCmd creates a new k3d command
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
	cmd.Flags().StringSliceVarP(
		&o.K3sArgs, "k3s-arg", "s", []string{},
		"One or more arguments passed from k3d to the k3s command (format: ARG@NODEFILTER[;@NODEFILTER])",
	)
	cmd.Flags().DurationVar(
		&o.Timeout, "timeout", 5*time.Minute, `Maximum time for the provisioning. If you want no timeout, enter "0".`,
	)
	cmd.Flags().StringSliceVarP(
		&o.K3dArgs, "k3d-arg", "", []string{},
		"One or more arguments passed to the k3d provisioning command (e.g. --k3d-arg='--no-rollback')",
	)
	cmd.Flags().StringSliceVarP(
		&o.K3dRegistryArgs, "k3d-registry-arg", "", []string{},
		"One or more arguments passed to the k3d registry create command (e.g. --k3d-registry-arg='--default-network podman')",
	)
	cmd.Flags().StringVarP(
		&o.KubernetesVersion, "kube-version", "k", provision.DefaultK8sFullVersion, "Kubernetes version of the cluster",
	)
	cmd.Flags().StringSliceVar(
		&o.UseRegistry, "registry-use", []string{},
		"Connect to one or more k3d-managed registries. Kyma automatically creates a registry for Serverless images.",
	)
	cmd.Flags().StringVar(
		&o.RegistryPort, "registry-port", "5001", "Specify the port on which the k3d registry will be exposed",
	)
	cmd.Flags().StringSliceVarP(
		&o.PortMapping, "port", "p", []string{"80:80@loadbalancer", "443:443@loadbalancer"},
		"Map ports 80 and 443 of K3D loadbalancer (e.g. -p 80:80@loadbalancer -p 443:443@loadbalancer)",
	)
	return cmd
}

// Run runs the command
func (c *command) Run() error {
	if c.opts.CI {
		c.Factory.NonInteractive = true
	}
	if c.opts.Verbose {
		c.Factory.UseLogger = true
	}

	var err error

	k3dClient := k3d.NewClient(k3d.NewCmdRunner(), k3d.NewPathLooker(), c.opts.Name, c.opts.Verbose, c.opts.Timeout)

	kinfo, err := c.initK3d(k3dClient)
	if err != nil {
		return err
	}

	if kinfo.manageDefaultRegistry {
		defaultRegistry, err := c.createK3dRegistry(k3dClient)
		if err != nil {
			return err
		}
		c.opts.UseRegistry = append(c.opts.UseRegistry, defaultRegistry)
	}

	return c.createK3dCluster(k3dClient)
}

// initK3d ensures that k3d is properly installed and pre-conditions are fulfilled
func (c *command) initK3d(k3dClient k3d.Client) (*k3dInfo, error) {
	vs := c.NewStep("Verifying k3d status")

	vs.LogInfo("Checking if port flags are valid")
	portsConfig, err := extractPortsFromFlag(c.opts.PortMapping)
	if err != nil {
		vs.Failure()
		return nil, errors.Wrapf(err, "Could not extract host ports from %s", c.opts.PortMapping)
	}

	kinfo, err := c.getK3dInfo(k3dClient)
	if err != nil {
		vs.Failure()
		return nil, err
	}

	err = c.cleanupK3d(k3dClient, kinfo, portsConfig)
	if err != nil {
		vs.Failure()
		return nil, err
	}
	vs.Successf("k3d status verified")
	return kinfo, nil
}

func (c *command) getK3dInfo(k3dClient k3d.Client) (*k3dInfo, error) {

	err := k3dClient.VerifyStatus()
	if err != nil {
		return nil, err
	}

	useDefaultRegistry := len(c.opts.UseRegistry) == 0
	var defaultRegistryExists bool

	if useDefaultRegistry {
		c.CurrentStep.LogInfo("Checking if k3d registry of previous kyma installation exists")
		defaultRegistryExists, err = k3dClient.RegistryExists()
		if err != nil {
			return nil, err
		}
	}

	c.CurrentStep.LogInfo("Checking if k3d cluster of previous kyma installation exists")
	clusterExists, err := k3dClient.ClusterExists()
	if err != nil {
		return nil, err
	}

	return &k3dInfo{
		clusterExists,
		useDefaultRegistry,
		defaultRegistryExists,
	}, nil
}

func (c *command) cleanupK3d(k3dClient k3d.Client, kinfo *k3dInfo, portsConfig []int) error {

	deleteRegistryIfRequired := func() error {
		if kinfo.manageDefaultRegistry && kinfo.defaultRegistryExists {
			if err := k3dClient.DeleteRegistry(); err != nil {
				return err
			}
			c.CurrentStep.LogInfo("Deleted k3d registry of previous kyma installation")
		}
		return nil
	}

	if kinfo.clusterExists {
		if !c.PromptUserToDeleteExistingCluster() {
			return fmt.Errorf("User decided not to remove the existing k3d cluster")
		}

		if err := deleteRegistryIfRequired(); err != nil {
			return err
		}

		if err := k3dClient.DeleteCluster(); err != nil {
			return err
		}
		c.CurrentStep.LogInfo("Deleted k3d cluster of previous kyma installation")

	} else {
		if err := allocatePorts(portsConfig...); err != nil {
			if strings.Contains(err.Error(), "bind: permission denied") {
				c.CurrentStep.LogInfo("Hint: The following error can potentially be mitigated by either running the command with `sudo` privileges or specifying other ports with the `--port` flag:")
			}
			return errors.Wrap(err, "Port cannot be allocated")
		}

		if err := deleteRegistryIfRequired(); err != nil {
			return err
		}
	}

	return nil
}

func (c *command) PromptUserToDeleteExistingCluster() bool {
	var answer bool
	if !c.opts.NonInteractive {
		answer = c.CurrentStep.PromptYesNo("Do you want to remove the existing k3d cluster? ")
	}
	return c.opts.NonInteractive || answer
}

func (c *command) createK3dRegistry(k3dClient k3d.Client) (string, error) {
	s := c.NewStep("Create k3d registry")

	registryURL, err := k3dClient.CreateRegistry(c.opts.RegistryPort, parseNestedArgs(c.opts.K3dRegistryArgs))
	if err != nil {
		s.Failuref("Could not create k3d registry")
		return "", err
	}
	s.Successf("Created k3d registry '%s'", registryURL)
	return registryURL, nil
}

// Create a k3d cluster
func (c *command) createK3dCluster(k3dClient k3d.Client) error {
	s := c.NewStep(fmt.Sprintf("Create K3d cluster '%s'", c.opts.Name))

	settings := k3d.CreateClusterSettings{
		Args:              parseNestedArgs(c.opts.K3dArgs),
		KubernetesVersion: c.opts.KubernetesVersion,
		PortMapping:       c.opts.PortMapping,
		Workers:           c.opts.Workers,
		K3sArgs:           c.opts.K3sArgs,
		UseRegistry:       c.opts.UseRegistry,
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

// Check if a port is allocated
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

func parseNestedArgs(args []string) []string {
	var res []string
	for _, arg := range args {
		res = append(res, strings.Split(arg, " ")...)
	}
	return res
}

type k3dInfo struct {
	clusterExists bool
	//indicates if the default k3d registry should be created/deleted
	manageDefaultRegistry bool
	//only valid if shouldManageDefaultRegistry is true
	defaultRegistryExists bool
}
