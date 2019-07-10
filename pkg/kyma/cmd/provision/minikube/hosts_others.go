// +build !windows

package minikube

import (
	"bytes"
	"fmt"
	"github.com/kyma-project/cli/internal/root"
	"os/exec"

	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/internal/step"
)

func addDevDomainsToEtcHostsOSSpecific(domain string, s step.Step, hostAlias string) error {
	notifyUserFunc := func(err error) {
		if err != nil {
			s.LogInfof("Error: %s", err.Error())
		}
		s.LogInfof("Execute the following command manually to add domain entries:\n###\n sudo sed -i.bak \"/"+domain+"/d\" "+hostsFile+" && echo '%s' | sudo tee -a /etc/hosts\r\n###\n", hostAlias)
	}

	s.LogInfo("Adding domain mappings to your 'hosts' file")
	if root.IsWithSudo() {
		s.LogInfo("You're running CLI with sudo. CLI has to add Kyma domain entries to your 'hosts'. Type 'y' to allow this action")
		if !root.PromptUser() {
			notifyUserFunc(nil)
			return nil
		}
	}
	_, err := internal.RunCmd("sudo",
		"sed", "-i.bak",
		fmt.Sprintf("/%s/d", domain),
		hostsFile)
	if err != nil {
		notifyUserFunc(err)
		return nil
	}

	cmd := exec.Command("sudo", "tee", "-a", hostsFile)
	buf := &bytes.Buffer{}
	_, err = fmt.Fprint(buf, hostAlias)
	if err != nil {
		notifyUserFunc(err)
		return nil
	}
	cmd.Stdin = buf
	err = cmd.Run()
	if err != nil {
		notifyUserFunc(err)
		return nil
	}
	return nil
}
