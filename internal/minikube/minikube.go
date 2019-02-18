package minikube

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/fsouza/go-dockerclient"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	minikubeVersion string = "0.33.0"
)

//RunCmd executes a minikube command with given arguments
func RunCmd(args ...string) (string, error) {
	cmd := exec.Command("minikube", args...)
	out, err := cmd.CombinedOutput()
	unquotedOut := strings.Replace(string(out), "'", "", -1)
	if err != nil {
		return unquotedOut, fmt.Errorf("Failed executing minikube command 'minikube %s' with output '%s' and error message '%s'", args, out, err)
	}
	return unquotedOut, nil
}

//CheckVersion checks whether minikube version is supported
func CheckVersion() (bool, error) {
	versionText, err := RunCmd("version")
	if err != nil {
		return false, err
	}

	exp, _ := regexp.Compile("minikube version: v(.*)")
	versionString := exp.FindStringSubmatch(versionText)
	version, err := semver.NewVersion(versionString[1])
	if err != nil {
		return false, err
	}

	constraintString := "~"+minikubeVersion
	constraint, err := semver.NewConstraint(constraintString)
	if err != nil {
		return false, err
	}

	return constraint.Check(version), nil
}

func DockerClient() (*docker.Client, error) {
	envOut, err := RunCmd("docker-env")
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
			envParts := strings.SplitN(env, "=",2)
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
