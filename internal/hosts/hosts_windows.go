//go:build windows
// +build windows

package hosts

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/pkg/step"
)

func addDevDomainsToEtcHostsOSSpecific(domain string, s step.Step, hostAlias string) error {

	err := addDevDomainsRunCmd(domain, s, hostAlias)
	if err != nil {
		return err
	}
	return nil
}

func addDevDomainsRunCmd(domain string, s step.Step, hostAlias string) error {

	hostsArray := strings.Split(hostAlias, " ")
	ip := hostsArray[0]
	hostsArray = hostsArray[1:]
	for len(hostsArray) > 0 {
		chunkLen := 7 // max hosts per line
		if len(hostsArray) < chunkLen {
			chunkLen = len(hostsArray)
		}

		addLine := ip + " " + strings.Join(hostsArray[:chunkLen], " ")
		hostsArray = hostsArray[chunkLen:]
		commandToRun := "echo " + addLine + " >> " + hostsFile
		_, err := cli.RunCmd("cmd", "/c", commandToRun)
		if err != nil {
			s.LogErrorf("Add these lines to your " + hostsFile + " file:")
			fmt.Printf("%s %s\n", ip, strings.Join(hostsArray[:chunkLen], " "))
		}
	}
	return nil
}
