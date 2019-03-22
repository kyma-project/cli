package internal

import (
	"strings"
	"time"

	"github.com/kyma-incubator/kyma-cli/internal/kubectl"
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
