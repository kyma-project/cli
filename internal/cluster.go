package internal

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/clebs/ldng"
)

const (
	sleep = 5 * time.Second
)

// RunCmd executes a command with given arguments
func RunCmd(c string, args []string) (string, error) {
	cmd := exec.Command(c, args[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed executing command '%s %s' with output '%s' and error message '%s'", c, args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}

//RunKubectlCmd executes a kubectl command with given arguments
func RunKubectlCmd(args []string) (string, error) {
	cmd := exec.Command("kubectl", args[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed executing kubectl command 'kubectl %s' with output '%s' and error message '%s'", args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}

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

//WaitForPod waits till a pod is deployed and has status 'running'.
// The pod gets identified by the namespace and a lebel key=value pair.
func WaitForPod(namespace string, labelName string, labelValue string) error {
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
	resourceNames, err := RunKubectlCmd(getResourceNameCmd)
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
	resourceNames, err := RunKubectlCmd(getResourceNameCmd)
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
	podNames, err := RunKubectlCmd(getPodNameCmd)
	if err != nil {
		return false, err
	}

	if podNames == "" {
		return false, nil
	}

	for _, pod := range strings.Split(podNames, "\\n") {
		getContainerStatusCmd := []string{"get", "pods", "-n", namespace, "-l", labelName + "=" + labelValue, "-o", "jsonpath='{.items[*].status.containerStatuses[0].ready}'"}
		containerStatus, err := RunKubectlCmd(getContainerStatusCmd)
		if err != nil {
			return false, err
		}

		if containerStatus != "true" {
			getEventsCmd := []string{"get", "event", "-n", namespace, "-o", "go-template='{{range .items}}{{if eq .involvedObject.name \"'" + pod + "'\"}}{{.message}}{{\"\\n\"}}{{end}}{{end}}'"}
			events, err := RunKubectlCmd(getEventsCmd)
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

//NewSpinner starts and returns a new terminal spinner using given text as progress and success text
func NewSpinner(startText string, stopText string) chan struct{} {
	s := ldng.NewSpin(ldng.SpinPrefix(fmt.Sprintf("%s", startText)), ldng.SpinPeriod(100*time.Millisecond), ldng.SpinSuccess(fmt.Sprintf("%s ✅\n", stopText)))
	return s.Start()
}

//StopSpinner marks a terminal spinner as finished with success
func StopSpinner(spinner chan struct{}) {
	spinner <- struct{}{} // stop the spinner
	<-spinner             // wait for the spinner to finish cleanup
	spinner = nil
}
