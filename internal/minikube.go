package internal

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"github.com/fsouza/go-dockerclient"
)

const (
	minikubeVersion string = "0.31.0"
)

//RunMinikubeCmd executes a minikube command with given arguments
func RunMinikubeCmd(args []string) (string, error) {
	cmd := exec.Command("minikube", args[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed executing minikube command 'minikube %s' with output '%s' and error message '%s'", args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}

//RunMinikubeCmdE executes a minikube command with given arguments ignoring any errors
func RunMinikubeCmdE(args []string) (string, error) {
	cmd := exec.Command("minikube", args[0:]...)
	out, _ := cmd.CombinedOutput()
	return strings.Replace(string(out), "'", "", -1), nil
}

//CheckMinikubeVersion assures that the minikube version used is compatible
func CheckMinikubeVersion() error {
	versionCmd := []string{"version"}
	versionText, err := RunMinikubeCmd(versionCmd)
	if err != nil {
		return err
	}

	exp, _ := regexp.Compile("minikube version: v((\\d+.\\d+.\\d+))")
	version := exp.FindStringSubmatch(versionText)

	if version[1] != minikubeVersion {
		return fmt.Errorf("Your minikube version is '%s'. Currently only minikube in version '%s' is supported", version[1], minikubeVersion)
	}
	return nil
}

func MinikubeDockerClient() (*docker.Client, error) {
	envOut, err := RunMinikubeCmd([]string{"docker-env"})
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
