// +build !windows

package minikube

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/internal/step"
)

func isWithSudo() bool {
	return os.Getenv("SUDO_UID") != ""
}

func promptUser() bool {
	for {
		fmt.Print("Type [y/n]: ")
		var res string
		if _, err := fmt.Scanf("%s", &res); err != nil {
			return false
		}
		switch res {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			continue
		}
	}
}

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, s step.Step, hostAlias string) error {
	notifyUserFunc := func(err error) {
		s.LogInfof("Error: %s. Execute the following command manually to add mappings: sudo sed -i '' \"/"+o.Domain+"/d\" "+hostsFile+" && echo '%s' | sudo tee -a /etc/hosts\r\n", err.Error(), hostAlias)
		return
	}

	s.LogInfo("Adding aliases to your 'hosts' file")
	if isWithSudo() {
		s.LogInfo("You're running CLI with sudo. CLI has to add Kyma domain entries to your 'hosts'. Type 'y' to allow this action")
		if !promptUser() {
			notifyUserFunc(nil)
			return nil
		}
	}
	_, err := internal.RunCmd("sudo", []string{
		"sed", "-i",
		"''",
		fmt.Sprintf("/%s/d", o.Domain),
		hostsFile})
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
