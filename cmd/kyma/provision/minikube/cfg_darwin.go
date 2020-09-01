// +build darwin

package minikube

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

const defaultVMDriver = vmDriverHyperkit
const allowVPNSock = true

func osSpecificFlags(o *Options, cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.UseVPNKitSock, "use-hyperkit-vpnkit-sock", false, `Uses vpnkit sock provided by Docker. This is useful when DNS Port (53) is being used by some other program like dns-proxy (eg. provided by Cisco Umbrella).`)
}

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
