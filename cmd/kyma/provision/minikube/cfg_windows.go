// +build windows

package minikube

import "github.com/spf13/cobra"

const defaultVMDriver = vmDriverVirtualBox
const allowVPNSock = false

func osSpecificFlags(o *Options, cmd *cobra.Command) {
}

func osSpecificRun(c *command, startCmd []string) ([]string, error) {
	return startCmd, nil
}
