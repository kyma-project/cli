// +build windows

package minikube

import (
	"fmt"

	"github.com/kyma-incubator/kyma-cli/internal/step"
)

const hostsFile = "C:\\Windows\\system32\\drivers\\etc\\hosts"

const defaultVMDriver = vmDriverVirtualBox

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, s step.Step, hostAlias string) error {

	s.LogErrorf("Please add these lines to your " + hostsFile + " file:")
	fmt.Println(hostAlias)

	return nil
}
