// +build !windows

package minikube

import (
	"github.com/kyma-incubator/kyma-cli/internal"
)

const HostsFile = "/etc/hosts"

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, hostAlias string) error {
	_, err := internal.RunCmd("sudo", []string{"/bin/sh", "-c", "sed -i '' \"/" + o.Domain + "/d\" " + HostsFile})
	if err != nil {
		return err
	}

	_, err = internal.RunCmd("sudo", []string{"/bin/sh", "-c", "echo '" + hostAlias + "' >> " + HostsFile})
	if err != nil {
		return err
	}

	return nil
}
