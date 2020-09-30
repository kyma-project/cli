// +build !windows
// +build !darwin

package minikube

import (
	"fmt"

	"github.com/spf13/cobra"
)

const defaultVMDriver = vmDriverNone

func osSpecificFlags(o *Options, cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.UseVPNKitSock, "use-hyperkit-vpnkit-sock", false, `Uses vpnkit sock provided by Docker. This is useful when DNS Port (53) is being used by some other program like dns-proxy (eg. provided by Cisco Umbrella. It works only on mac OS).`)
}

func osSpecificRun(c *command, startCmd []string) ([]string, error) {
	if c.opts.UseVPNKitSock {
		fmt.Println("This flag is only supported for Mac OS!")
	}
	return startCmd, nil
}
