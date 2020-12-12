package k3s

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	docker "github.com/docker/docker/client"

	"github.com/blang/semver/v4"
)

const (
	k3sVersion string = "1.19.0"
)

//RunCmd executes a minikube command with given arguments
func RunCmd(verbose bool, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "k3d", args...)

	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)

	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("Executing 'k3d %s' command with output '%s' timed out, try running the command manually or increasing timeout using the 'timeout' flag", strings.Join(args, " "), out)
	}

	if err != nil {
		if verbose {
			fmt.Printf("\nExecuted command:\n  k3d %s\nwith output:\n  %s\nand error:\n  %s\n", strings.Join(args, " "), string(out), err)
		}
		return out, fmt.Errorf("Executing the 'k3d %s' command with output '%s' and error message '%s' failed", strings.Join(args, " "), out, err)
	}
	if verbose {
		fmt.Printf("\nExecuted command:\n  k3d %s\nwith output:\n  %s\n", strings.Join(args, " "), string(out))
	}
	return out, nil
}

//CheckVersion checks whether minikube version is supported
func CheckVersion(verbose bool, timeout time.Duration) (string, error) {
	versionOutput, err := RunCmd(verbose, timeout, "version")
	if err != nil {
		return "", err
	}

	exp, _ := regexp.Compile("k3s version v([^\\s-]+)")
	versionString := exp.FindStringSubmatch(versionOutput)
	if verbose {
		fmt.Printf("Extracted K3s version: '%s'", versionString[1])
	}
	version, err := semver.Parse(versionString[1])
	if err != nil {
		return "", err
	}

	supportedVersion, _ := semver.Parse(k3sVersion)
	if version.LT(supportedVersion) {
		return "", fmt.Errorf("You are using an unsupported k3s version '%s'. "+
			"This may not work. The recommended k3s version is '%s'", version, supportedVersion)
	}

	return "", nil
}

//DockerClient creates a docker client
func DockerClient(verbose bool, timeout time.Duration) (*docker.Client, error) {
	dockerClient, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}

	return dockerClient, nil
}
