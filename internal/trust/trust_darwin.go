// +build darwin

package trust

import (
	"fmt"

	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/internal/root"
	"github.com/pkg/errors"
)

type keychain struct{}

func NewCertifier() Certifier {
	return keychain{}
}

func (k keychain) StoreCertificate(file string) error {
	fmt.Println("Kyma needs to add its certificates to the keychain...")
	if root.IsWithSudo() {
		fmt.Println("You're running CLI with sudo. CLI has to add the Kyma certificate to the keychain. Type 'y' to allow this action.")
		if !root.PromptUser() {
			fmt.Println("Opertion aborted")
			return nil
		}
	}

	_, err := internal.RunCmd("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", file)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the kyma certificates, please follow the instructions below to import them manually:\n%s", k.Instructions()))
	}

	return nil
}

func (keychain) Instructions() string {
	return "1. Download the certificate: kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\\.ingress\\.tlsCrt}' | base64 --decode > kyma.crt\n" +
		"2. Import the certificate: sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt\n"
}
