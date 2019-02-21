package internal

import (
	"strings"

	"github.com/kyma-incubator/kyma-cli/internal/kubectl"
)

//GetKymaVersion determines the version of kyma installed to current cluster
func GetKymaVersion() (string, error) {
	kymaVersion, err := kubectl.RunCmd([]string{"-n", "kyma-installer", "get", "pod", "-l", "name=kyma-installer", "-o", "jsonpath='{.items[*].spec.containers[0].image}'"})
	if err != nil {
		return "", err
	}
	if kymaVersion == "" {
		return "N/A", nil
	}
	return strings.Split(kymaVersion, ":")[1], nil
}
