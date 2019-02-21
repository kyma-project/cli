package minikube

import (
	"fmt"
)

const HostsFile = "C:\\Windows\\system32\\drivers\\etc\\hosts"

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, hostAlias string) error {
	fmt.Println()
	fmt.Println("=====")
	fmt.Println("Please add these lines to your " + HostsFile + " file:")
	fmt.Println(hostAlias)
	fmt.Println("=====")

	return nil
}
