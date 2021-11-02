package k3d

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
)

const (
	v4MinRequiredVersion         string = "4.0.0"
	v5MinRequiredVersion         string = "5.0.0"
	v5DefaultRegistryNamePattern string = "%s-registry"

	binaryName string = "k3d"
)

type Client interface {
	runCmd(args ...string) (string, error)
	checkVersion(minRequiredVersion string) error
	VerifyStatus(isV5 bool) error
	ClusterExists() (bool, error)
	RegistryExists() (bool, error)
	CreateCluster(settings CreateClusterSettings, isV5 bool) error
	CreateRegistry() (string, error)
	DeleteCluster() error
	DeleteRegistry() error
}

type CreateClusterSettings struct {
	Args              []string
	KubernetesVersion string
	PortMap           map[string]int
	PortMapping       []string
	Workers           int
	V4Settings        V4CreateClusterSettings
	V5Settings        V5CreateClusterSettings
}

type V4CreateClusterSettings struct {
	ServerArgs []string
	AgentArgs  []string
}

type V5CreateClusterSettings struct {
	K3sArgs     []string
	UseRegistry []string
}

type client struct {
	cmdRunner   CmdRunner
	pathLooker  PathLooker
	clusterName string
	verbose     bool
	userTimeout time.Duration
}

// NewClient creates a new instance of the Client interface.
// The 'isAlpha' parameter indicates whether the command is an alpha command. If so, the minimum k3d version is v5.
func NewClient(cmdRunner CmdRunner, pathLooker PathLooker, clusterName string, verbose bool, timeout time.Duration) Client {
	return &client{
		cmdRunner:   cmdRunner,
		pathLooker:  pathLooker,
		clusterName: clusterName,
		verbose:     verbose,
		userTimeout: timeout,
	}
}

func (c *client) runCmd(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.userTimeout)
	defer cancel()

	out, err := c.cmdRunner.Run(ctx, binaryName, args...)

	if err != nil {
		if c.verbose {
			fmt.Printf("Failing command:\n  %s %s\nwith output:\n  %s\nand error:\n  %s\n", binaryName, strings.Join(args, " "), out, err)
		}
		return out, errors.Wrapf(err, "Executing '%s %s' failed with output '%s'", binaryName, strings.Join(args, " "), out)
	}

	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("Executing '%s %s' command with output '%s' timed out. Try running the command manually or increasing the timeout using the 'timeout' flag", binaryName, strings.Join(args, " "), out)
	}

	if c.verbose {
		fmt.Printf("\nExecuted command:\n %s %s\nwith output:\n  %s\n", binaryName, strings.Join(args, " "), out)
	}
	return out, nil
}

//checkVersion checks whether k3d version is supported
func (c *client) checkVersion(minRequiredVersion string) error {
	binaryVersionOutput, err := c.runCmd("version")
	if err != nil {
		return err
	}

	exp, _ := regexp.Compile(fmt.Sprintf(`%s version v([^\s-]+)`, binaryName))
	binaryVersion := exp.FindStringSubmatch(binaryVersionOutput)
	if c.verbose {
		fmt.Printf("Extracted %s version: '%s'", binaryName, binaryVersion[1])
	}
	if len(binaryVersion) < 2 {
		return fmt.Errorf("Could not extract %s version from command output:\n%s", binaryName, binaryVersionOutput)
	}
	binarySemVersion, err := semver.Parse(binaryVersion[1])
	if err != nil {
		return err
	}

	minRequiredSemVersion, _ := semver.Parse(minRequiredVersion)
	if binarySemVersion.Major > minRequiredSemVersion.Major {
		incompatibleMajorVersionMsg := "You are using an unsupported k3d major version '%d'. The supported k3d major version for this command is '%d'."
		if minRequiredVersion == v4MinRequiredVersion {
			incompatibleMajorVersionMsg += "\n\nIf you want to use k3d v5, try the dedicated 'kyma alpha provision k3d' command"
		}
		return fmt.Errorf(incompatibleMajorVersionMsg, binarySemVersion.Major, minRequiredSemVersion.Major)
	} else if binarySemVersion.LT(minRequiredSemVersion) {
		incompatibleVersionMsg := "You are using an unsupported k3d version '%s'. The supported k3d version for this command is >= '%s'."
		return fmt.Errorf(incompatibleVersionMsg, binaryVersion, minRequiredSemVersion)
	}

	return nil
}

//getRegistryByName gets one k3d registry by name
func (c *client) getRegistryByName(registryName string) (*Registry, error) {
	registryJSON, err := c.runCmd("registry", "list", "-o", "json")
	if err != nil {
		return nil, err
	}

	registryList := &RegistryList{}
	if err := registryList.Unmarshal([]byte(registryJSON)); err != nil {
		return nil, err
	}

	for _, registry := range registryList.Registries {
		if registry.Name == fmt.Sprintf("k3d-%s", registryName) {
			if c.verbose {
				fmt.Printf("k3d registry '%s' exists", registryName)
			}
			return &registry, nil
		}
	}

	if c.verbose {
		fmt.Printf("k3d registry '%s' does not exist", registryName)
	}
	return nil, nil
}

//VerifyStatus verifies whether the k3d CLI tool is properly installed
func (c *client) VerifyStatus(isV5 bool) error {
	//ensure k3d is in PATH
	if _, err := c.pathLooker.Look(binaryName); err != nil {
		if c.verbose {
			fmt.Printf("Command '%s' not found in PATH", binaryName)
		}
		return fmt.Errorf("Command '%s' not found. Please install %s (see https://github.com/rancher/k3d#get)", binaryName, binaryName)
	}

	var minRequiredVersion string
	if isV5 {
		minRequiredVersion = v5MinRequiredVersion
	} else {
		minRequiredVersion = v4MinRequiredVersion
	}
	if err := c.checkVersion(minRequiredVersion); err != nil {
		return err
	}

	// execute a command and return the error
	_, err := c.runCmd("cluster", "list")
	return err
}

//ClusterExists checks whether a cluster exists
func (c *client) ClusterExists() (bool, error) {
	clusterJSON, err := c.runCmd("cluster", "list", "-o", "json")
	if err != nil {
		return false, err
	}

	clusterList := &ClusterList{}
	if err := clusterList.Unmarshal([]byte(clusterJSON)); err != nil {
		return false, err
	}

	for _, cluster := range clusterList.Clusters {
		if cluster.Name == c.clusterName {
			if c.verbose {
				fmt.Printf("k3d cluster '%s' exists", c.clusterName)
			}
			return true, nil
		}
	}

	if c.verbose {
		fmt.Printf("k3d cluster '%s' does not exist", c.clusterName)
	}
	return false, nil
}

//RegistryExists checks whether a registry exists
func (c *client) RegistryExists() (bool, error) {
	registryName := fmt.Sprintf(v5DefaultRegistryNamePattern, c.clusterName)

	registry, err := c.getRegistryByName(registryName)
	if err != nil {
		return false, err
	}
	return registry != nil, nil
}

//CreateCluster creates a cluster
func (c *client) CreateCluster(settings CreateClusterSettings, isV5 bool) error {
	k3sImage, err := getK3sImage(settings.KubernetesVersion)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"cluster", "create", c.clusterName,
		"--kubeconfig-update-default",
		"--timeout", fmt.Sprintf("%ds", int(c.userTimeout.Seconds())),
		"--agents", fmt.Sprintf("%d", settings.Workers),
		"--image", k3sImage,
	}

	if isV5 {
		cmdArgs = append(cmdArgs, getCreateClusterV5Args(settings)...)
	} else {
		cmdArgs = append(cmdArgs, getCreateClusterV4Args(settings)...)
	}

	cmdArgs = append(cmdArgs, constructArgs("--port", settings.PortMapping)...)
	//add further k3d args which are not offered by the Kyma CLI flags
	cmdArgs = append(cmdArgs, settings.Args...)

	_, err = c.runCmd(cmdArgs...)
	return err
}

// CreateRegistry creates a k3d registry
func (c *client) CreateRegistry() (string, error) {
	registryName := fmt.Sprintf(v5DefaultRegistryNamePattern, c.clusterName)
	registryPort := "5001"

	_, err := c.runCmd("registry", "create", registryName, "--port", registryPort)

	return fmt.Sprintf("%s:%s", registryName, registryPort), err
}

// DeleteCluster deletes a k3d registry
func (c *client) DeleteCluster() error {
	_, err := c.runCmd("cluster", "delete", c.clusterName)
	return err
}

// DeleteRegistry deletes a k3d registry
func (c *client) DeleteRegistry() error {
	registryName := fmt.Sprintf(v5DefaultRegistryNamePattern, c.clusterName)

	_, err := c.runCmd("registry", "delete", fmt.Sprintf("k3d-%s", registryName))
	return err
}

func getCreateClusterV4Args(settings CreateClusterSettings) []string {
	cmdArgs := []string{
		"--registry-create",
		"--k3s-server-arg", "--disable",
		"--k3s-server-arg", "traefik",
	}
	cmdArgs = append(cmdArgs, constructArgs("--k3s-server-arg", settings.V4Settings.ServerArgs)...)
	cmdArgs = append(cmdArgs, constructArgs("--k3s-agent-arg", settings.V4Settings.AgentArgs)...)

	return cmdArgs
}

func getCreateClusterV5Args(settings CreateClusterSettings) []string {
	cmdArgs := []string{
		"--kubeconfig-switch-context",
		"--k3s-arg", "--disable=traefik@server:0",
	}
	cmdArgs = append(cmdArgs, constructArgs("--registry-use", settings.V5Settings.UseRegistry)...)
	cmdArgs = append(cmdArgs, constructArgs("--k3s-arg", settings.V5Settings.K3sArgs)...)

	return cmdArgs
}

func getK3sImage(kubernetesVersion string) (string, error) {
	_, err := semver.Parse(kubernetesVersion)
	if err != nil {
		return "", fmt.Errorf("Invalid Kubernetes version %v: %v", kubernetesVersion, err)
	}

	return fmt.Sprintf("rancher/k3s:v%s-k3s1", kubernetesVersion), nil
}

func constructArgs(argName string, rawPorts []string) []string {
	var portMap []string
	for _, port := range rawPorts {
		portMap = append(portMap, argName, port)
	}
	return portMap
}
