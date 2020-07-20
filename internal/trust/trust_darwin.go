// +build darwin

package trust

import (
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"

	"github.com/kyma-project/cli/internal/root"
	"github.com/pkg/errors"
)

type keychain struct {
	k8s    kube.KymaKube
	source Source
}

func NewCertifier(k kube.KymaKube, src Source) Certifier {
	return keychain{
		k8s:    k,
		source: src,
	}
}
func (k keychain) Certificate() ([]byte, error) {
	if k.source.resource == CertSourceSecret {
		return certificateFromSecret(k)
	}
	return certificateFromConfigMap(k)
}

func (k keychain) StoreCertificate(file string, i Informer) error {
	i.LogInfo("Kyma wants to add its root certificate to the keychain.")
	if root.IsWithSudo() {
		i.LogInfo("You're running CLI with sudo. CLI has to add the Kyma certificate to the keychain. Type 'y' to allow this action.")
		if !root.PromptUser() {
			i.LogInfo(fmt.Sprintf("\nCould not import the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", k.Instructions()))
			return nil
		}
	}

	_, err := cli.RunCmd("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", file)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the Kyma root certificate. Follow the instructions below to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	return nil
}

func (keychain) Instructions() string {
	return "1. Download the certificate: kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\\.ingress\\.tlsCrt}' | base64 --decode > kyma.crt\n" +
		"2. Import the certificate: sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt\n"
}
