// +build darwin

package minikube

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/internal/step"
)

const hostsFile = "/etc/hosts"

const defaultVMDriver = vmDriverHyperkit

func isWithSudo() bool {
	val := os.Getenv("SUDO_UID")
	if val != "" {
		return true
	}
	return false
}

func promptUser() bool {
	for {
		fmt.Print("Type [y/n]: ")
		var res string
		_, err := fmt.Scanf("%s", &res)
		if err != nil {
			return false
		}
		switch res {
		case "yes":
			return true
		case "no":
			return false
		default:
			continue
		}
	}
}

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, s step.Step, hostAlias string) error {
	notifyUserFunc := func() {
		s.LogInfof("Execute the following command manually to add mappings: sudo sed -i '' \"/"+o.Domain+"/d\" "+hostsFile+" && echo '%s' | sudo tee -a /etc/hosts\r\n", hostAlias)
		return
	}

	s.LogInfo("Adding aliases to your 'hosts' file")
	if isWithSudo() {
		s.LogInfo("You're running CLI with sudo. CLI has to add Kyma domain entries to your 'hosts'. Type 'y' to allow this action")
		if !promptUser() {
			notifyUserFunc()
			return nil
		}
	}
	_, err := internal.RunCmd("sudo", []string{"/bin/sh", "-c", "sed -i '' \"/" + o.Domain + "/d\" " + hostsFile})
	if err != nil {
		notifyUserFunc()
		return nil
	}

	cmd := exec.Command("sudo", "tee", "-a", hostsFile)
	buf := &bytes.Buffer{}
	_, err = fmt.Fprint(buf, hostAlias)
	if err != nil {
		notifyUserFunc()
		return nil
	}
	cmd.Stdin = buf
	err = cmd.Run()
	if err != nil {
		notifyUserFunc()
		return nil
	}
	return nil
}
