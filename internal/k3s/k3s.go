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
	k3dMinVersion  string        = "1.19.0"
	defaultTimeout time.Duration = 10 * time.Second
)

//RunCmd executes a minikube command with given arguments
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

//CheckVersion checks whether minikube version is supported
func CheckVersion(verbose bool) error {
	versionOutput, err := RunCmd(verbose, defaultTimeout, "version")
	if err != nil {
		return err
	}

	exp, _ := regexp.Compile(`k3s version v([^\s-]+)`)
	versionString := exp.FindStringSubmatch(versionOutput)
	if verbose {
		fmt.Printf("Extracted K3s version: '%s'", versionString[1])
	}
	version, err := semver.Parse(versionString[1])
	if err != nil {
		return err
	}

	minVersion, _ := semver.Parse(k3dMinVersion)
	if version.LT(minVersion) {
		return fmt.Errorf("You are using an unsupported k3s version '%s'. "+
			"This may not work. The recommended k3s version is '%s'", version, minVersion)
	}

	return nil
}

//Initialize verifies whether the k3d CLI tool is properly installed
func Initialize(verbose bool) error {
	//ensure k3d is in PATH
	if _, err := exec.LookPath("k3d"); err != nil {
		if verbose {
			fmt.Printf("Command 'k3d' not found in PATH")
		}
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

//StartCluster starts a cluster
func StartCluster(verbose bool, timeout time.Duration, clusterName string, workers int) error {
	_, err := RunCmd(verbose, timeout,
		"cluster", "create", clusterName,
		"--timeout", fmt.Sprintf("%ds", int(timeout.Seconds())),
		"-p", "80:80@loadbalancer", "-p", "443:443@loadbalancer",
		"--k3s-server-arg", "--no-deploy", "--k3s-server-arg", "traefik",
		"--switch-context",
		"--agents", fmt.Sprintf("%d", workers),
	)
	return err
}

//DeleteCluster deletes a cluster
func DeleteCluster(verbose bool, timeout time.Duration, clusterName string) error {
	_, err := RunCmd(verbose, timeout, "cluster", "delete", clusterName)
	return err
}
