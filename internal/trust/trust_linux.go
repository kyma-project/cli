//go:build linux
// +build linux

package trust

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/root"
)

type certauth struct {
	k8s kube.KymaKube
}

func NewCertifier(k kube.KymaKube) Certifier {
	return certauth{
		k8s: k,
	}
}

func (c certauth) Certificate() ([]byte, error) {
	cm, err := c.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Get(context.Background(), "net-global-overrides", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", c.Instructions()))
	}

	decodedCert, err := base64.StdEncoding.DecodeString(cm.Data["global.ingress.tlsCrt"])
	if err != nil {
		return nil, err
	}

	return decodedCert, nil
}

func (c certauth) CertificateKyma2() ([]byte, error) {
	s, err := c.k8s.Static().CoreV1().Secrets("istio-system").Get(context.Background(), "kyma-gateway-certs", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", c.InstructionsKyma2()))
	}

	return s.Data["tls.crt"], nil
}

func (c certauth) StoreCertificate(file string, i Informer) error {
	i.LogInfo("Kyma wants to add its root certificate to the trusted certificate import.")
	if er := root.IsWithSudo(); er == nil {
		i.LogInfo("You're running CLI with sudo. CLI has to add the Kyma certificate to the trusted certificate import. Type 'y' to allow this action.")
		if !root.PromptUser() {
			i.LogInfo(fmt.Sprintf("\nCould not import the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", c.Instructions()))
			return nil
		}
	}

	// get domain to put on the certificate name.
	// Linux does not have a proper certificate manager and we need to be able to identify the certificate
	domain, err := certDomain(file)
	if err != nil {
		return err
	}

	_, err = cli.RunCmd("sudo", "cp", file, fmt.Sprintf("/usr/local/share/ca-certificates/kyma-%s.crt", domain))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the Kyma certificates. Follow the instructions to import them manually:\n-----\n%s-----\n", c.Instructions()))
	}
	_, err = cli.RunCmd("sudo", "update-ca-certificates")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the Kyma certificates. Follow the instructions to import them manually:\n-----\n%s-----\n", c.Instructions()))
	}

	return nil
}

func (certauth) Instructions() string {
	return "1. Download the certificate: kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\\.ingress\\.tlsCrt}' | base64 --decode > kyma.crt\n" +
		"2. Rename the certificate file: mv kyma.crt {NEW_CERT_NAME}\n" +
		"3. Copy the certificate to the CA folder: sudo cp {NEW_CERT_NAME} /usr/local/share/ca-certificates/\n" +
		"4. Update the certificate registry: sudo update-ca-certificates\n"
}

func (certauth) InstructionsKyma2() string {
	return "1. Download the certificate: kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\\.crt}' > kyma.crt\n" +
		"2. Rename the certificate file: mv kyma.crt {NEW_CERT_NAME}\n" +
		"3. Copy the certificate to the CA folder: sudo cp {NEW_CERT_NAME} /usr/local/share/ca-certificates/\n" +
		"4. Update the certificate registry: sudo update-ca-certificates\n"
}

// certDomain returns the DNS info of the provided root certificate.
func certDomain(certFile string) (string, error) {
	certText, err := cli.RunCmd("openssl", "x509", "-text", "-noout", "-in", certFile)
	if err != nil {
		return "", err
	}

	matches := regexp.MustCompile("DNS:(.*)[\r\n]+").FindStringSubmatch(certText)

	if len(matches) < 2 {
		return "", errors.New("Could not determine the certificate's DNS")
	}

	return strings.Replace(matches[1], "'", "", -1), nil
}

func (c certauth) StoreCertificateKyma2(file string, _ Informer) error {

	// get domain to put on the certificate name.
	// Linux does not have a proper certificate manager and we need to be able to identify the certificate
	domain, err := certDomain(file)
	if err != nil {
		return err
	}

	_, err = cli.RunCmd("sudo", "cp", file, fmt.Sprintf("/usr/local/share/ca-certificates/kyma-%s.crt", domain))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the Kyma certificates. Follow the instructions to import them manually:\n-----\n%s-----\n", c.InstructionsKyma2()))
	}
	_, err = cli.RunCmd("sudo", "update-ca-certificates")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the Kyma certificates. Follow the instructions to import them manually:\n-----\n%s-----\n", c.InstructionsKyma2()))
	}

	return nil
}
