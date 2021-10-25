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
	V4MinVersion                 string = "4.0.0"
	V5MinVersion                 string = "5.0.0"
	V5DefaultRegistryNamePattern string = "%s-registry"

	binaryName string = "k3d"
)

type Client interface {
	runCmd(args ...string) (string, error)
	checkVersion() error
	VerifyStatus() error
	ClusterExists() (bool, error)
	RegistryExists(registryName string) (bool, error)
	CreateCluster(settings CreateClusterSettings) error
	CreateRegistry(registryName string) (string, error)
	DeleteCluster() error
	DeleteRegistry(registryName string) error
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
	minVersion  string
	verbose     bool
	userTimeout time.Duration
}

// NewClient creates a new instance of the Client interface.
// The 'isAlpha' parameter indicates whether the command is an alpha command. If so, the minimum k3d version is v5.
func NewClient(cmdRunner CmdRunner, pathLooker PathLooker, clusterName string, verbose bool, timeout time.Duration, isAlpha bool) Client {
	var mink3dVersion string
	if isAlpha {
		mink3dVersion = V5MinVersion
	} else {
		mink3dVersion = V4MinVersion
	}

	return &client{
		cmdRunner:   cmdRunner,
		pathLooker:  pathLooker,
		clusterName: clusterName,
		verbose:     verbose,
		userTimeout: timeout,
		minVersion:  mink3dVersion,
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
func (c *client) checkVersion() error {
	versionOutput, err := c.runCmd("version")
	if err != nil {
		return err
	}

	exp, _ := regexp.Compile(fmt.Sprintf(`%s version v([^\s-]+)`, binaryName))
	versionString := exp.FindStringSubmatch(versionOutput)
	if c.verbose {
		fmt.Printf("Extracted %s version: '%s'", binaryName, versionString[1])
	}
	if len(versionString) < 2 {
		return fmt.Errorf("Could not extract %s version from command output:\n%s", binaryName, versionOutput)
	}
	version, err := semver.Parse(versionString[1])
	if err != nil {
		return err
	}

	minVersion, _ := semver.Parse(c.minVersion)
	if version.Major > minVersion.Major {
		return fmt.Errorf("You are using an unsupported %s major version '%d' for this command. "+
			"This may not work. The recommended %s major version for this command is '%d'", binaryName, version.Major, binaryName, minVersion.Major)
	} else if version.LT(minVersion) {
		return fmt.Errorf("You are using an unsupported %s version '%s' for this command. "+
			"This may not work. The recommended %s version for this command is >= '%s'", binaryName, version, binaryName, minVersion)
	}

	return nil
}

//VerifyStatus verifies whether the k3d CLI tool is properly installed
func (c *client) VerifyStatus() error {
	//ensure k3d is in PATH
	if _, err := c.pathLooker.Look(binaryName); err != nil {
		if c.verbose {
			fmt.Printf("Command '%s' not found in PATH", binaryName)
		}
		return fmt.Errorf("Command '%s' not found. Please install %s (see https://github.com/rancher/k3d#get)", binaryName, binaryName)
	}

	if err := c.checkVersion(); err != nil {
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
func (c *client) RegistryExists(registryName string) (bool, error) {
	registryJSON, err := c.runCmd("registry", "list", "-o", "json")
	if err != nil {
		return false, err
	}

	registryList := &RegistryList{}
	if err := registryList.Unmarshal([]byte(registryJSON)); err != nil {
		return false, err
	}

	for _, registry := range registryList.Registries {
		if registry.Name == fmt.Sprintf("k3d-%s", registryName) {
			if c.verbose {
				fmt.Printf("k3d registry '%s' exists", registryName)
			}
			return true, nil
		}
	}

	if c.verbose {
		fmt.Printf("k3d registry '%s' does not exist", registryName)
	}
	return false, nil
}

//CreateCluster creates a cluster
func (c *client) CreateCluster(settings CreateClusterSettings) error {
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

	if c.minVersion == V4MinVersion {
		cmdArgs = append(cmdArgs, getCreateClusterV4Args(settings)...)
	} else {
		cmdArgs = append(cmdArgs, getCreateClusterV5Args(settings)...)
	}

	cmdArgs = append(cmdArgs, constructArgs("--port", settings.PortMapping)...)
	//add further k3d args which are not offered by the Kyma CLI flags
	cmdArgs = append(cmdArgs, settings.Args...)

	_, err = c.runCmd(cmdArgs...)
	return err
}

// CreateRegistry creates a k3d registry
func (c *client) CreateRegistry(registryName string) (string, error) {
	registryPort := "5000"

	_, err := c.runCmd("registry", "create", registryName, "--port", registryPort)
	return fmt.Sprintf("%s:%s", registryName, registryPort), err
}

// DeleteCluster deletes a k3d registry
func (c *client) DeleteCluster() error {
	_, err := c.runCmd("cluster", "delete", c.clusterName)
	return err
}

// DeleteRegistry deletes a k3d registry
func (c *client) DeleteRegistry(registryName string) error {
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
