// +build darwin

package minikube

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli/internal/cli"
)

const defaultVMDriver = vmDriverHyperkit

func osSpecificRun(c *command, startCmd []string) ([]string, error) {
	if c.opts.UseVPNKitSock {
		user, err := cli.RunCmd("whoami")
		if err != nil {
			return nil, err
		}
		pathToVPNKitSock := fmt.Sprintf("/Users/%s/Library/Containers/com.docker.docker/Data/vpnkit.eth.sock", strings.TrimSuffix(user, "\n"))
		startCmd = append(startCmd, "--hyperkit-vpnkit-sock="+pathToVPNKitSock)
	}
	return startCmd, nil
}
