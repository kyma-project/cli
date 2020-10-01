// +build windows

package minikube

import "github.com/kyma-project/cli/internal/nice"

const defaultVMDriver = vmDriverVirtualBox

func osSpecificRun(c *command, startCmd []string) ([]string, error) {
	if c.opts.UseVPNKitSock {
		np := nice.Nice{}
		np.PrintImportant("WARNING: This flag is supported only for Mac OS!")
	}
	return startCmd, nil
}
