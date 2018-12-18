package internal

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const (
	minikubeVersion string = "0.28.2"
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
