package minikube

import (
	"fmt"
	"github.com/kyma-incubator/kyma-cli/internal"
)

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, hostAlias string) error {
	fmt.Println()
	fmt.Println("=====")
	fmt.Println("Please add these lines to your " + internal.HOSTS_FILE + " file:")
	fmt.Println(hostAlias)
	fmt.Println("=====")

	return nil
}
