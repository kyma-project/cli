// +build !windows

package minikube

import (
	"github.com/kyma-incubator/kyma-cli/internal"
)

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, hostAlias string) error {
	_, err := internal.RunCmd("sudo", []string{"/bin/sh", "-c", "sed -i '' \"/" + o.Domain + "/d\" " + internal.HOSTS_FILE})
	if err != nil {
		return err
	}

	_, err = internal.RunCmd("sudo", []string{"/bin/sh", "-c", "echo '" + hostAlias + "' >> " + internal.HOSTS_FILE})
	if err != nil {
		return err
	}

	return nil
}