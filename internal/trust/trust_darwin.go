// +build darwin

package trust

import (
	"encoding/base64"
	"fmt"

	"github.com/kyma-project/cli/internal/step"

	"github.com/kyma-project/cli/internal/kubectl"

	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/internal/root"
	"github.com/pkg/errors"
)

type keychain struct {
	kubectl *kubectl.Wrapper
	step    step.Step
}

func NewCertifier(verbose bool) Certifier {
	return keychain{
		kubectl: kubectl.NewWrapper(verbose),
	}
}

func (k keychain) Certificate() ([]byte, error) {
	cert, err := k.kubectl.RunCmd("get", "configmap", "net-global-overrides", "-n", "kyma-installer", "-o", "jsonpath='{.data.global\\.ingress\\.tlsCrt}'")
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not obtain the Kyma root certificate, please follow the instructions below to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	decodedCert, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not obtain the Kyma root certificate, please follow the instructions below to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	return decodedCert, nil
}

func (k keychain) StoreCertificate(file string, i Informer) error {
	i.LogInfo("Kyma wants to add its root certificate to the keychain.")
	if root.IsWithSudo() {
		i.LogInfo("You're running CLI with sudo. CLI has to add the Kyma certificate to the keychain. Type 'y' to allow this action.")
		if !root.PromptUser() {
			i.LogInfo(fmt.Sprintf("\nCould not import the Kyma root certificate, please follow the instructions below to import it manually:\n-----\n%s-----\n", k.Instructions()))
			return nil
		}
	}

	_, err := internal.RunCmd("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", file)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the Kyma root certificate, please follow the instructions below to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	return nil
}

func (keychain) Instructions() string {
	return "1. Download the certificate: kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\\.ingress\\.tlsCrt}' | base64 --decode > kyma.crt\n" +
		"2. Import the certificate: sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt\n"
}
