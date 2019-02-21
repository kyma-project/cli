package kubectl

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	kubectlVersion string = "1.10.0"
	sleep                 = 5 * time.Second
)

//RunCmd executes a kubectl command with given arguments
func RunCmd(args []string) (string, error) {
	cmd := exec.Command("kubectl", args[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed executing kubectl command 'kubectl %s' with output '%s' and error message '%s'", args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}

//WaitForPodReady waits till a pod is deployed and has status 'running'.
// The pod gets identified by the namespace and a lebel key=value pair.
func WaitForPodReady(namespace string, labelName string, labelValue string) error {
	for {
		isDeployed, err := IsPodDeployed(namespace, labelName, labelValue)
		if err != nil {
			return err
		}
		if isDeployed {
			break
		}
		time.Sleep(sleep)
	}

	for {
		isReady, err := IsPodReady(namespace, labelName, labelValue)
		if err != nil {
			return err
		}
		if isReady {
			break
		}
		time.Sleep(sleep)
	}
	return nil
}

//WaitForPodGone waits till a pod is not existent anymore.
// The pod gets identified by the namespace and a lebel key=value pair.
func WaitForPodGone(namespace string, labelName string, labelValue string) error {
	for {
		check, err := IsPodDeployed(namespace, labelName, labelValue)
		if err != nil {
			return err
		}
		if !check {
			break
		}
		time.Sleep(sleep)
	}
	return nil
}

//IsPodDeployed checks if a pod is deployed.
// It will not wait till it is deployed.
// The pod gets identified by the namespace and a label key=value pair.
func IsPodDeployed(namespace string, labelName string, labelValue string) (bool, error) {
	return IsResourceDeployed("pod", namespace, labelName, labelValue)
}

//IsResourceDeployed checks if a kubernetes resource is deployed.
// It will not wait till it is deployed.
// The resource gets identified by the namespace and a lebel key=value pair.
func IsResourceDeployed(resource string, namespace string, labelName string, labelValue string) (bool, error) {
	getResourceNameCmd := []string{"get", resource, "-n", namespace, "-l", labelName + "=" + labelValue, "-o", "jsonpath='{.items[*].metadata.name}'"}
	resourceNames, err := RunCmd(getResourceNameCmd)
	if err != nil {
		return false, err
	}
	if resourceNames == "" {
		return false, nil
	}
	return true, nil
}

//IsClusterResourceDeployed checks if a kubernetes cluster resource is deployed.
// It will not wait till it is deployed.
// The resource gets identified by a lebel key=value pair.
func IsClusterResourceDeployed(resource string, labelName string, labelValue string) (bool, error) {
	getResourceNameCmd := []string{"get", resource, "-l", labelName + "=" + labelValue, "-o", "jsonpath='{.items[*].metadata.name}'"}
	resourceNames, err := RunCmd(getResourceNameCmd)
	if err != nil {
		return false, err
	}
	if resourceNames == "" {
		return false, nil
	}
	return true, nil
}

//IsPodReady checks if a pod is deployed and running.
// It will not wait till it is deployed or running.
// The pod gets identified by the namespace and a lebel key=value pair.
func IsPodReady(namespace string, labelName string, labelValue string) (bool, error) {
	getPodNameCmd := []string{"get", "pods", "-n", namespace, "-l", labelName + "=" + labelValue, "-o", "jsonpath='{.items[*].metadata.name}'"}
	podNames, err := RunCmd(getPodNameCmd)
	if err != nil {
		return false, err
	}

	if podNames == "" {
		return false, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(podNames))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}

		pod := scanner.Text()
		getContainerStatusCmd := []string{"get", "pod", pod, "-n", namespace, "-o", "jsonpath='{.status.containerStatuses[0].ready}'"}
		containerStatus, err := RunCmd(getContainerStatusCmd)
		if err != nil {
			return false, err
		}

		if containerStatus != "true" {
			getEventsCmd := []string{"get", "event", "-n", namespace, "-o", "go-template='{{range .items}}{{if eq .involvedObject.name \"'" + pod + "'\"}}{{.message}}{{\"\\n\"}}{{end}}{{end}}'"}
			events, err := RunCmd(getEventsCmd)
			if err != nil {
				fmt.Printf("Error while checking for pod events '%s'\n‚", err)
			}
			if events != "" {
				fmt.Printf("Status '%s'", events)
			}
			return false, nil
		}
	}
	return true, nil
}

//CheckVersion assures that the kubectl version used is compatible
func CheckVersion() (string, error) {
	versionText, err := RunCmd([]string{"version", "--client", "--short"})
	if err != nil {
		return "", err
	}

	exp, _ := regexp.Compile("Client Version: v((\\d+).(\\d+).(\\d+))")
	kubctlIsVersion := exp.FindStringSubmatch(versionText)

	exp, _ = regexp.Compile("((\\d+).(\\d+).(\\d+))")
	kubctlMustVersion := exp.FindStringSubmatch(kubectlVersion)

	majorIsVersion, _ := strconv.Atoi(kubctlIsVersion[2])
	majorMustVersion, _ := strconv.Atoi(kubctlMustVersion[2])
	minorIsVersion, _ := strconv.Atoi(kubctlIsVersion[3])
	minorMustVersion, _ := strconv.Atoi(kubctlMustVersion[3])

	message := fmt.Sprintf("Your kubectl version is '%s'. Supported versions of kubectl are from '%d.%d.*' to '%d.%d.*'", kubctlIsVersion[1], majorMustVersion, minorMustVersion-1, majorMustVersion, minorMustVersion+1)

	if majorIsVersion != majorMustVersion {
		return "", errors.New(message)
	}
	if minorIsVersion-minorMustVersion < -1 || minorIsVersion-minorMustVersion > 1 {
		return message, nil
	}
	return "", nil
}
