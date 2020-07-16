package k3d

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	docker "github.com/docker/docker/client"

	"github.com/Masterminds/semver"
)

const (
	k3dVersion string = "1.7.0"
)

//RunCmd executes a k3d command with given arguments
func RunCmd(timeout time.Duration, rawArgs ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	args := []string{}
	args = append(args, rawArgs...)

	cmd := exec.CommandContext(ctx, "k3d", args...)

	out, err := cmd.CombinedOutput()
	unquotedOut := strings.Replace(string(out), "'", "", -1)

	if ctx.Err() == context.DeadlineExceeded {
		return unquotedOut, fmt.Errorf("executing 'k3d %s' command with output '%s' timed out, try running the command manually or increasing timeout using the 'timeout' flag", strings.Join(args, " "), out)
	}

	if err != nil {
		return unquotedOut, fmt.Errorf("Executing the 'k3d %s' command with output '%s' and error message '%s' failed", strings.Join(args, " "), out, err)
	}
	return unquotedOut, nil
}

//CheckVersion checks whether k3d version is supported
func CheckVersion(verbose bool, timeout time.Duration) (string, error) {
	versionText, err := RunCmd(timeout, "version")
	if err != nil {
		return "", err
	}

	exp, _ := regexp.Compile("k3d version: v(.*)")
	versionString := exp.FindStringSubmatch(versionText)
	version, err := semver.NewVersion(versionString[1])
	if err != nil {
		return "", err
	}

	constraintString := "~" + k3dVersion
	constraint, err := semver.NewConstraint(constraintString)
	if err != nil {
		return "", err
	}

	check := constraint.Check(version)
	if check {
		return "", nil
	}
	return fmt.Sprintf("You are using an unsupported k3d version '%s'. This may not work. The recommended k3d version is '%s'", version, k3dVersion), nil
}

//DockerClient creates a docker client based on local host env
func DockerClient() (*docker.Client, error) {
	dc, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}

	return dc, nil
}
