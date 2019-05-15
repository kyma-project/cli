// +build !windows
// +build !darwin

package minikube

import (
	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/internal/step"
)

const hostsFile = "/etc/hosts"

const defaultVMDriver = vmDriverNone

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, s step.Step, hostAlias string) error {
	s.LogInfo("Please enter your password if requested")

	_, err := internal.RunCmd("sudo", []string{"/bin/sh", "-c", "sed -i '' \"/" + o.Domain + "/d\" " + hostsFile})
	if err != nil {
		return err
	}

	_, err = internal.RunCmd("sudo", []string{"/bin/sh", "-c", "echo '" + hostAlias + "' >> " + hostsFile})
	if err != nil {
		return err
	}

	return nil
}
