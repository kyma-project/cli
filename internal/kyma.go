package internal

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/kubectl"
)

//GetKymaVersion determines the version of kyma installed to current cluster
func GetKymaVersion(verbose bool) (string, error) {
	kymaVersion, err := kubectl.RunCmdWithTimeout(2*time.Second, verbose, "-n", "kyma-installer", "get", "pod", "-l", "name=kyma-installer", "-o", "jsonpath='{.items[*].spec.containers[0].image}'")
	if err != nil {
		return "", err
	}
	if kymaVersion == "" {
		return "N/A", nil
	}
	return strings.Split(kymaVersion, ":")[1], nil
}

// RunCmd executes a command with given arguments
func RunCmd(c string, args ...string) (string, error) {
	cmd := exec.Command(c, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed executing command '%s %s' with output '%s' and error message '%s'", c, args, out, err)
	}
	return strings.Replace(string(out), "'", "", -1), nil
}
