package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/clebs/ldng"
)

const (
	sleep = 5 * time.Second
)

func RunKubeCmd(c []string) string {
	cmd := exec.Command("kubectl", c[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing kubectl command: 'kubectl %s'\n", c)
		fmt.Printf("Error message is: %s\n", out)
		os.Exit(1)
	}
	return strings.Replace(string(out), "'", "", -1)
}

func RunMinikubeCmd(c []string) string {
	cmd := exec.Command("minikube", c[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing minikube command: 'minikube %s'\n", c)
		fmt.Printf("Error message is: %s\n", out)
		os.Exit(1)
	}
	return strings.Replace(string(out), "'", "", -1)
}

func IsPodReady(namespace string, labelName string, labelValue string) string {

	for {
		isDeployed := isDeployed(namespace, labelName, labelValue)
		if isDeployed {
			break
		}
		time.Sleep(sleep)
	}

	for {
		isReady := isReady(namespace, labelName, labelValue)
		if isReady {
			break
		}
		time.Sleep(sleep)
	}
	return ""
}

func isDeployed(namespace string, labelName string, labelValue string) bool {
	getPodNameCmd := []string{"get", "pods", "-n", namespace, "-l", labelName + "=" + labelValue, "-o", "jsonpath='{.items[*].metadata.name}'"}
	podNames := RunKubeCmd(getPodNameCmd)
	if podNames == "" {
		return false
	}
	return true
}

func isReady(namespace string, labelName string, labelValue string) bool {
	getPodNameCmd := []string{"get", "pods", "-n", namespace, "-l", labelName + "=" + labelValue, "-o", "jsonpath='{.items[*].metadata.name}'"}
	podNames := RunKubeCmd(getPodNameCmd)
	for _, pod := range strings.Split(podNames, "\\n") {
		getContainerStatusCmd := []string{"get", "pods", "-n", namespace, "-l", labelName + "=" + labelValue, "-o", "jsonpath='{.items[*].status.containerStatuses[0].ready}'"}
		containerStatus := RunKubeCmd(getContainerStatusCmd)

		if containerStatus != "true" {
			getEventsCmd := []string{"get", "event", "-n", namespace, "-o", "go-template='{{range .items}}{{if eq .involvedObject.name \"'" + pod + "'\"}}{{.message}}{{\"\\n\"}}{{end}}{{end}}'"}
			events := RunKubeCmd(getEventsCmd)
			if events != "" {
				fmt.Printf("Status '%s'", events)
			}
			return false
		}
	}
	return true
}

func NewSpinner(startText string, stopText string) chan struct{} {
	s := ldng.NewSpin(ldng.SpinPrefix(fmt.Sprintf("%s", startText)), ldng.SpinPeriod(100*time.Millisecond), ldng.SpinSuccess(fmt.Sprintf("%s ✅\n", stopText)))
	return s.Start()
}

func StopSpinner(spinner chan struct{}) {
	spinner <- struct{}{} // stop the spinner
	<-spinner             // wait for the spinner to finish cleanup
	spinner = nil
}
