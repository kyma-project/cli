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

func addDevDomainsToEtcHostsOSSpecific(o *MinikubeOptions, s step.Step, hostAlias string) error {
	s.LogInfo("Adding aliases to your 'hosts' file")
	_, err := internal.RunCmd("sudo", []string{"/bin/sh", "-c", "sed -i '' \"/" + o.Domain + "/d\" " + hostsFile})
	if err != nil {
		return err
	}
	if isWithSudo() {
		s.LogInfo("You're running CLI with sudo. CLI has to add dns mapping to your 'hosts'. Type 'yes' to allow this action")
		for {
			fmt.Print("Type 'yes' or 'no': ")
			var res string
			_, err := fmt.Scanf("%s", &res)
			if err != nil {
				continue
			}
			switch res {
			case "yes":
				cmd := exec.Command("sudo", "tee", "-a", hostsFile)
				buf := &bytes.Buffer{}
				_, err = fmt.Fprint(buf, hostAlias)
				if err != nil {
					return err
				}
				cmd.Stdin = buf
				err = cmd.Run()
				if err != nil {
					return err
				}
				return nil
			case "no":
				break
			default:
				continue
			}
		}
	}
	fmt.Printf("Execute the following command manually to add mappings: echo '%s' | tee -a /etc/hosts\r\n", hostAlias)

	return nil
}
