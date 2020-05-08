package minikube

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	docker "github.com/fsouza/go-dockerclient"
)

const (
	minikubeVersion string = "1.6.2"
)

//RunCmd executes a minikube command with given arguments
func RunCmd(verbose bool, profile string, timeout time.Duration, rawArgs ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{}
	if profile != "" {
		args = append(args, "--profile")
		args = append(args, profile)
	}
	args = append(args, rawArgs...)

	cmd := exec.CommandContext(ctx, "minikube", args...)

	out, err := cmd.CombinedOutput()
	unquotedOut := strings.Replace(string(out), "'", "", -1)

	if ctx.Err() == context.DeadlineExceeded {
		return unquotedOut, fmt.Errorf("Executing 'minikube %s' command with output '%s' timed out, try running the command manually or increasing timeout using 'timeout' flag", strings.Join(args, " "), out)
	}

	if err != nil {
		if verbose {
			fmt.Printf("\nExecuted command:\n  minikube %s\nwith output:\n  %s\nand error:\n  %s\n", strings.Join(args, " "), string(out), err)
		}
		return unquotedOut, fmt.Errorf("Executing the 'minikube %s' command with output '%s' and error message '%s' failed", strings.Join(args, " "), out, err)
	}
	if verbose {
		fmt.Printf("\nExecuted command:\n  minikube %s\nwith output:\n  %s\n", strings.Join(args, " "), string(out))
	}
	return unquotedOut, nil
}

//CheckVersion checks whether minikube version is supported
func CheckVersion(verbose bool, timeout time.Duration) (string, error) {
	versionText, err := RunCmd(verbose, "", timeout, "version")
	if err != nil {
		return "", err
	}

	exp, _ := regexp.Compile("minikube version: v(.*)")
	versionString := exp.FindStringSubmatch(versionText)
	version, err := semver.NewVersion(versionString[1])
	if err != nil {
		return "", err
	}

	constraintString := "~" + minikubeVersion
	constraint, err := semver.NewConstraint(constraintString)
	if err != nil {
		return "", err
	}

	check := constraint.Check(version)
	if check {
		return "", nil
	}
	return fmt.Sprintf("You are using an unsupported Minikube version '%s'. This may not work. The recommended Minikube version is '%s'", version, minikubeVersion), nil
}

//DockerClient creates a docker client based on minikube "docker-env" configuration
func DockerClient(verbose bool, profile string, timeout time.Duration) (*docker.Client, error) {
	envOut, err := RunCmd(verbose, profile, timeout, "docker-env", "--shell", "bash")
	if err != nil {
		return nil, err
	}

	oldEnvs := make(map[string]string)
	defer func() {
		for key, val := range oldEnvs {
			os.Setenv(key, val)
		}
	}()
	for _, line := range strings.Split(envOut, "\n") {
		if strings.HasPrefix(line, "export") {
			env := strings.SplitN(line, " ", 2)[1]
			envParts := strings.SplitN(env, "=", 2)
			envKey := envParts[0]
			envVal := strings.Trim(envParts[1], `"`)
			oldEnvs[envKey] = os.Getenv(envKey)
			err := os.Setenv(envKey, envVal)
			if err != nil {
				return nil, err
			}
		}
	}
	return docker.NewClientFromEnv()
}
