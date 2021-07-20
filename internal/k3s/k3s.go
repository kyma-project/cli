package k3s

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
)

const (
	k3dMinVersion  string        = "4.0.0"
	defaultTimeout time.Duration = 10 * time.Second
)

//RunCmd executes a k3d command with given arguments
func RunCmd(verbose bool, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "k3d", args...)

	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	if err != nil {
		if verbose {
			fmt.Printf("Failing command:\n  k3d %s\nwith output:\n  %s\nand error:\n  %s\n", strings.Join(args, " "), string(out), err)
		}
		return out, errors.Wrapf(err, "Executing 'k3d %s' failed with output '%s'", strings.Join(args, " "), out)
	}

	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("Executing 'k3d %s' command with output '%s' timed out, try running the command manually or increasing timeout using the 'timeout' flag", strings.Join(args, " "), out)
	}

	if verbose {
		fmt.Printf("\nExecuted command:\n  k3d %s\nwith output:\n  %s\n", strings.Join(args, " "), string(out))
	}
	return out, nil
}

//checkVersion checks whether k3s version is supported
func checkVersion(verbose bool) error {
	versionOutput, err := RunCmd(verbose, defaultTimeout, "version")
	if err != nil {
		return err
	}

	exp, _ := regexp.Compile(`k3d version v([^\s-]+)`)
	versionString := exp.FindStringSubmatch(versionOutput)
	if verbose {
		fmt.Printf("Extracted K3d version: '%s'", versionString[1])
	}
	if len(versionString) < 2 {
		return fmt.Errorf("Could not extract k3d version from command output:\n%s", versionOutput)
	}
	version, err := semver.Parse(versionString[1])
	if err != nil {
		return err
	}

	minVersion, _ := semver.Parse(k3dMinVersion)
	if version.Major > minVersion.Major {
		return fmt.Errorf("You are using an unsupported k3d major version '%d'. "+
			"This may not work. The recommended k3d major version is '%d'", version.Major, minVersion.Major)
	} else if version.LT(minVersion) {
		return fmt.Errorf("You are using an unsupported k3d version '%s'. "+
			"This may not work. The recommended k3d version is >= '%s'", version, minVersion)
	}

	return nil
}

func getK3sImage(version string) (string, error) {
	_, err := semver.Parse(version)
	if err != nil {
		return "", fmt.Errorf("Invalid Kubernetes version %v: %v", version, err)
	}

	return fmt.Sprintf("rancher/k3s:v%s-k3s1", version), nil
}

//Initialize verifies whether the k3d CLI tool is properly installed
func Initialize(verbose bool) error {
	//ensure k3d is in PATH
	if _, err := exec.LookPath("k3d"); err != nil {
		if verbose {
			fmt.Printf("Command 'k3d' not found in PATH")
		}
		return fmt.Errorf("Command 'k3d' not found. Please install k3d following the installation " +
			"instructions provided at https://github.com/rancher/k3d#get")
	}

	if err := checkVersion(verbose); err != nil {
		return err
	}

	//verify whether k3d seems to be properly installed
	_, err := RunCmd(verbose, defaultTimeout, "cluster", "list")
	return err
}

//ClusterExists checks whether a cluster exists
func ClusterExists(verbose bool, clusterName string) (bool, error) {
	args := []string{"cluster", "list", "-o", "json"}
	clusterJSON, err := RunCmd(verbose, defaultTimeout, args...)
	if err != nil {
		return false, err
	}
	if verbose {
		fmt.Printf("K3d cluster list JSON: '%s'", clusterJSON)
	}

	clusterList := &ClusterList{}
	if err := clusterList.Unmarshal([]byte(clusterJSON)); err != nil {
		return false, err
	}

	for _, cluster := range clusterList.Clusters {
		if cluster.Name == clusterName {
			if verbose {
				fmt.Printf("K3d cluster '%s' exists", clusterName)
			}
			return true, nil
		}
	}

	if verbose {
		fmt.Printf("K3d cluster '%s' does not exist", clusterName)
	}
	return false, nil
}

func constructArgs(argname string, rawPorts []string) []string {
	portMap := []string{}
	for _, port := range rawPorts {
		portMap = append(portMap, argname, port)
	}
	return portMap
}

type Settings struct {
	ClusterName string
	Args        []string
	Version     string
	PortMap     map[string]int
	PortMapping []string
}

//StartCluster starts a cluster
func StartCluster(verbose bool, timeout time.Duration, workers int, serverArgs []string, agentArgs []string, k3d Settings) error {
	k3sImage, err := getK3sImage(k3d.Version)
	if err != nil {
		return err
	}
	cmdArgs := []string{
		"cluster", "create", k3d.ClusterName,
		"--kubeconfig-update-default",
		"--timeout", fmt.Sprintf("%ds", int(timeout.Seconds())),
		"--agents", fmt.Sprintf("%d", workers),
		"--registry-create",
		"--image", k3sImage,
		"--k3s-server-arg", "--disable",
		"--k3s-server-arg", "traefik",
	}

	cmdArgs = append(cmdArgs, constructArgs("--k3s-server-arg", serverArgs)...)
	cmdArgs = append(cmdArgs, constructArgs("--k3s-agent-arg", agentArgs)...)
	cmdArgs = append(cmdArgs, constructArgs("-p", k3d.PortMapping)...)

	//add further k3d args which are not offered by the Kyma CLI flags
	cmdArgs = append(cmdArgs, k3d.Args...)

	_, err = RunCmd(verbose, timeout, cmdArgs...)

	return err
}

//DeleteCluster deletes a cluster
func DeleteCluster(verbose bool, timeout time.Duration, clusterName string) error {
	_, err := RunCmd(verbose, timeout, "cluster", "delete", clusterName)
	return err
}
