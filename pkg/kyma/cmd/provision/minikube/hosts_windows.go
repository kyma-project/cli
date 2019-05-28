// +build windows

package minikube

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli/internal/step"
)

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, s step.Step, hostAlias string) error {

	s.LogErrorf("Please add these lines to your " + hostsFile + " file:")
	hostsArray := strings.Split(hostAlias, " ")
	ip := hostsArray[0]
	hostsPerLine := split(hostsArray[1:], 7) // 7 hostnames per line

	for k := 0; k < len(hostsPerLine); k++ {
		fmt.Printf("%s %s\n", ip, strings.Join(hostsPerLine[k], " "))
	}

	return nil
}

func split(buf []string, lim int) [][]string {
	var chunk []string
	chunks := make([][]string, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}
