package kubectl

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	yaml "gopkg.in/yaml.v2"
)

const (
	kubectlVersion string = "1.12.5"
	sleep                 = 5 * time.Second
)

//RunCmd executes a kubectl command with given arguments
func RunCmd(verbose bool, args ...string) (string, error) {
	cmd := exec.Command("kubectl", args[0:]...)
	return execCmd(cmd, strings.Join(args, " "), verbose)
}

func RunCmdWithTimeout(timeout time.Duration, verbose bool, args ...string) (string, error) {
	ctx, timeoutF := context.WithTimeout(context.Background(), timeout)
	defer timeoutF()
	cmd := exec.CommandContext(ctx, "kubectl", args[0:]...)
	result, err := execCmd(cmd, strings.Join(args, " "), verbose)
	if ctx.Err() != nil {
		result, err = "", ctx.Err()
	}
	return result, err
}

//RunApplyCmd executes a kubectl apply command with given resources
func RunApplyCmd(resources []map[string]interface{}, verbose bool) (string, error) {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	for _, y := range resources {
		err := enc.Encode(y)
		if err != nil {
			return "", err
		}
	}
	err := enc.Close()
	if err != nil {
		return "", err
	}
	cmd.Stdin = buf
	return execCmd(cmd, fmt.Sprintf("apply -f -%s", resources), verbose)
}

func execCmd(cmd *exec.Cmd, inputText string, verbose bool) (string, error) {
	out, err := cmd.CombinedOutput()
	unquotedOut := strings.Replace(string(out), "'", "", -1)
	if err != nil {
		if verbose {
			fmt.Printf("\nExecuted command:\n  kubectl %s\nwith output:\n  %s\nand error:\n  %s\n", inputText, string(out), err)
		}
		return unquotedOut, fmt.Errorf("Failed executing kubectl command 'kubectl %s' with output '%s' and error message '%s'", inputText, out, err)
	}
	if verbose {
		fmt.Printf("\nExecuted command:\n  kubectl %s\nwith output:\n  %s\n", inputText, string(out))
	}
	return unquotedOut, nil
}

//WaitForPodReady waits till a pod is deployed and has status 'running'.
// The pod gets identified by the namespace and a lebel key=value pair.
func WaitForPodReady(namespace string, labelName string, labelValue string, verbose bool) error {
	for {
		isDeployed, err := IsPodDeployed(namespace, labelName, labelValue, verbose)
		if err != nil {
			return err
		}
		if isDeployed {
			break
		}
		time.Sleep(sleep)
	}

	for {
		isReady, err := IsPodReady(namespace, labelName, labelValue, verbose)
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
func WaitForPodGone(namespace string, labelName string, labelValue string, verbose bool) error {
	for {
		check, err := IsPodDeployed(namespace, labelName, labelValue, verbose)
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
func IsPodDeployed(namespace string, labelName string, labelValue string, verbose bool) (bool, error) {
	return IsResourceDeployed("pod", namespace, labelName, labelValue, verbose)
}

//IsResourceDeployed checks if a kubernetes resource is deployed.
// It will not wait till it is deployed.
// The resource gets identified by the namespace and a lebel key=value pair.
func IsResourceDeployed(resource string, namespace string, labelName string, labelValue string, verbose bool) (bool, error) {
	resourceNames, err := RunCmd(verbose, "get", resource, "-n", namespace, "-l", labelName+"="+labelValue, "-o", "jsonpath='{.items[*].metadata.name}'")
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
func IsClusterResourceDeployed(resource string, labelName string, labelValue string, verbose bool) (bool, error) {
	resourceNames, err := RunCmd(verbose, "get", resource, "-l", labelName+"="+labelValue, "-o", "jsonpath='{.items[*].metadata.name}'")
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
func IsPodReady(namespace string, labelName string, labelValue string, verbose bool) (bool, error) {
	podNames, err := RunCmd(verbose, "get", "pods", "-n", namespace, "-l", labelName+"="+labelValue, "-o", "jsonpath='{.items[*].metadata.name}'")
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
		containerStatus, err := RunCmd(verbose, "get", "pod", pod, "-n", namespace, "-o", "jsonpath='{.status.containerStatuses[0].ready}'")
		if err != nil {
			return false, err
		}

		if containerStatus != "true" {
			events, err := RunCmd(verbose, "get", "event", "-n", namespace, "-o", "go-template='{{range .items}}{{if eq .involvedObject.name \"'"+pod+"'\"}}{{.message}}{{\"\\n\"}}{{end}}{{end}}'")
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
func CheckVersion(verbose bool) (string, error) {
	versionText, err := RunCmd(verbose, "version", "--client", "--short")
	if err != nil {
		return "", err
	}

	exp, _ := regexp.Compile("Client Version: v(.*)")
	versionString := exp.FindStringSubmatch(versionText)
	version, err := semver.NewVersion(versionString[1])
	if err != nil {
		return "", err
	}

	constraintString := "~" + kubectlVersion
	constraint, err := semver.NewConstraint(constraintString)
	if err != nil {
		return "", err
	}

	check := constraint.Check(version)
	if check {
		return "", nil
	}

	return fmt.Sprintf("You are using an unsupported kubectl version '%s'. This may not work. It is recommended to use kubectl version '%s'", version, kubectlVersion), nil
}
