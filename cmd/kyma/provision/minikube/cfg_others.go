// +build !windows
// +build !darwin

package minikube

import "github.com/spf13/cobra"

const defaultVMDriver = vmDriverNone
const allowVPNSock = false

func osSpecificFlags(o *Options, cmd *cobra.Command) {
}

func osSpecificRun(c *command, startCmd []string) ([]string, error) {
	return startCmd, nil
}
