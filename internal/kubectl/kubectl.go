package kubectl

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const (
	kubectlVersion string = "1.12.5"
	sleep                 = 5 * time.Second
)

//RunCmd executes a kubectl command with given arguments
func runCmd(verbose bool, args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	return execCmd(cmd, strings.Join(args, " "), verbose)
}

//RunApplyCmd executes a kubectl apply command with given resources
func runApplyCmd(resources []map[string]interface{}, verbose bool, kubeconfig string) (string, error) {
	cmd := exec.Command("kubectl", fmt.Sprintf("--kubeconfig=%s", kubeconfig), "apply", "-f", "-")
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
		return unquotedOut, fmt.Errorf("Failed executing kubectl 'kubectl %s' command  with output '%s' and error message '%s'", inputText, out, err)
	}
	if verbose {
		fmt.Printf("\nExecuted command:\n  kubectl %s\nwith output:\n  %s\n", inputText, string(out))
	}
	return unquotedOut, nil
}
