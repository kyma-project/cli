// +build windows

package install

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli/pkg/step"
)

func addDevDomainsToEtcHostsOSSpecific(domain string, s step.Step, hostAlias string) error {

	s.LogErrorf("Add these lines to your " + hostsFile + " file:")
	hostsArray := strings.Split(hostAlias, " ")
	ip := hostsArray[0]
	hostsArray = hostsArray[1:]
	for len(hostsArray) > 0 {
		chunkLen := 7 // max hosts per line
		if len(hostsArray) < chunkLen {
			chunkLen = len(hostsArray)
		}
		fmt.Printf("%s %s\n", ip, strings.Join(hostsArray[:chunkLen], " "))
		hostsArray = hostsArray[chunkLen:]
	}
	return nil
}
