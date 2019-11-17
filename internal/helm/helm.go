package helm

import (
	"os"
	"os/exec"
	"strings"

	// Initialize GCP client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var helmCmd = exec.Command("helm", "home")

// Home returns the path to the helm configuration folder or an error
func Home() (string, error) {
	helmHomeRaw, err := helmCmd.CombinedOutput()
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
