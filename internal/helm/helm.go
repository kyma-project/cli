package helm

import (
	"os"
	"os/exec"
	"strings"

	// Initialize GCP client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var helmVersionCmd = exec.Command("helm", "version", "--short", "--client")
var helmHomeCmd = exec.Command("helm", "home")

// SupportedVersion returns if the Helm version is supported by Kyma (currently v2.x.x)
func SupportedVersion() (bool, error) {
	helmVersionRaw, err := helmVersionCmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	if strings.Contains(string(helmVersionRaw), "v2") {
		return true, nil
	}

	return false, nil
}

// Home returns the path to the helm configuration folder or an error
func Home() (string, error) {
	helmHomeRaw, err := helmHomeCmd.CombinedOutput()
	if err != nil {
		return "", nil
	}

	helmHome := strings.Replace(string(helmHomeRaw), "\n", "", -1)
	if _, err := os.Stat(helmHome); os.IsNotExist(err) {
		err = os.MkdirAll(helmHome, 0700)
		if err != nil {
			return "", err
		}
	}
	return helmHome, nil
}
