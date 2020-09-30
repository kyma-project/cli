// +build windows

package minikube

import (
	"fmt"

	"github.com/spf13/cobra"
)

const defaultVMDriver = vmDriverVirtualBox

func osSpecificFlags(o *Options, cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.UseVPNKitSock, "use-hyperkit-vpnkit-sock", false, `Uses vpnkit sock provided by Docker. This is useful when DNS Port (53) is being used by some other program like dns-proxy (eg. provided by Cisco Umbrella.  This flag works only on Mac OS).`)
}

func osSpecificRun(c *command, startCmd []string) ([]string, error) {
	if c.opts.UseVPNKitSock {
		fmt.Println("Warning: This flag is only supported for Mac OS!")
	}
	return startCmd, nil
}
