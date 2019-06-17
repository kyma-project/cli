// +build windows

package trust

import (
	"fmt"

	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/internal/root"
	"github.com/pkg/errors"
)

type certutil struct{}

func NewCertifier() Certifier {
	return certutil{}
}

func (c certutil) StoreCertificate(file string) error {
	fmt.Println("Kyma needs to add its certificates to the trusted certificates...")

	if root.IsWithSudo() {
		fmt.Println("You're running CLI with sudo. CLI has to add the Kyma certificate to the trusted certificates. Type 'y' to allow this action.")
		if !root.PromptUser() {
			fmt.Println("Opertion aborted") // TODO change this
			return nil
		}
		// Only automatically add the cert if already on admin mode, can't ask for admin password from go
		_, err := internal.RunCmd("certutil", "-addstore", "-f", "Root", file)
		return err
	}

	return errors.New(fmt.Sprintf("Could not import the kyma certificates, please follow the instructions below to import them manually:\n%s", c.Instructions()))
}

func (certutil) Instructions() string {
	return "1. Open a new command line with administrator rights.\n" +
		"2. Download the certificate: kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\\.ingress\\.tlsCrt}' | base64 --decode > kyma.crt\n" +
		"3. Import the certificate: certutil -addstore -f Root kyma.crt\n"
}
