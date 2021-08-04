// +build darwin

package trust

import (
	"context"
	"encoding/base64"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"

	"github.com/kyma-project/cli/internal/root"
	"github.com/pkg/errors"
)

type keychain struct {
	k8s kube.KymaKube
}

func NewCertifier(k kube.KymaKube) Certifier {
	return keychain{
		k8s: k,
	}
}

func (k keychain) Certificate() ([]byte, error) {
	cm, err := k.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Get(context.Background(), "net-global-overrides", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	decodedCert, err := base64.StdEncoding.DecodeString(cm.Data["global.ingress.tlsCrt"])
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	return decodedCert, nil
}

func (k keychain) CertificateKyma2() ([]byte, error) {
	s, err := k.k8s.Static().CoreV1().Secrets("istio-system").Get(context.Background(), "kyma-gateway-certs", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	return s.Data["tls.crt"], nil
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

func (keychain) InstructionsKyma2() string {
	return "1. Download the certificate: kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\\.crt}' > kyma.crt\n" +
		"2. Import the certificate: sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt\n"
}
