package minikube

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	docker "github.com/fsouza/go-dockerclient"
)

const (
	minikubeVersion string = "0.33.0"
)

//RunCmd executes a minikube command with given arguments
func RunCmd(verbose bool, args ...string) (string, error) {
	cmd := exec.Command("minikube", args...)
	out, err := cmd.CombinedOutput()
	unquotedOut := strings.Replace(string(out), "'", "", -1)

	if err != nil {
		if verbose {
			fmt.Printf("\nExecuted command:\n  minikube %s\nwith output:\n  %s\nand error:\n  %s\n", strings.Join(args, " "), string(out), err)
		}
		return unquotedOut, fmt.Errorf("Failed executing minikube command 'minikube %s' with output '%s' and error message '%s'", args, out, err)
	}
	if verbose {
		fmt.Printf("\nExecuted minikube command:\n  %s\nwith output:\n  %s\n", args, string(out))
	}
	return unquotedOut, nil
}

//CheckVersion checks whether minikube version is supported
func CheckVersion(verbose bool) (string, error) {
	versionText, err := RunCmd(verbose, "version")
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
	return fmt.Sprintf("You are using an unsupported minikube version '%s'. This may not work. It is recommended to use minikube version '%s'", version, minikubeVersion), nil
}

//DockerClient creates a docker client based on minikube "docker-env" configuration
func DockerClient(verbose bool) (*docker.Client, error) {
	envOut, err := RunCmd(verbose, "docker-env")
	if err != nil {
		return nil, err
	}

	oldEnvs := make(map[string]string)
	defer func() {
		for key, val := range oldEnvs {
			_ = os.Setenv(key, val)
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
