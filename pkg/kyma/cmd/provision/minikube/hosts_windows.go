// +build windows

package minikube

import (
	"fmt"

	"github.com/kyma-project/cli/internal/step"
)

func addDevDomainsToEtcHostsOSSpecific(domain string, s step.Step, hostAlias string) error {

	s.LogErrorf("Please add these lines to your " + hostsFile + " file:")
	fmt.Println(hostAlias)

	return nil
}
