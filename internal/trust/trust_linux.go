// +build linux

package trust

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal"
	"github.com/kyma-project/cli/internal/kubectl"
	"github.com/kyma-project/cli/internal/root"
)

type certauth struct {
	kubectl *kubectl.Wrapper
}

func NewCertifier(verbose bool) Certifier {
	return certauth{
		kubectl: kubectl.NewWrapper(verbose),
	}
}

func (c certauth) Certificate() ([]byte, error) {
	cert, err := c.kubectl.RunCmd("get", "configmap", "net-global-overrides", "-n", "kyma-installer", "-o", "jsonpath='{.data.global\\.ingress\\.tlsCrt}'")
	if err != nil {
		return nil, err
	}

	decodedCert, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return nil, err
	}

	return decodedCert, nil
}

func (c certauth) StoreCertificate(file string, i Informer) error {
	i.LogInfo("Kyma wants to add its root certificate to the trusted certificate store.")
	if root.IsWithSudo() {
		i.LogInfo("You're running CLI with sudo. CLI has to add the Kyma certificate to the trusted certificate store. Type 'y' to allow this action.")
		if !root.PromptUser() {
			i.LogInfo(fmt.Sprintf("\nCould not import the Kyma root certificate, please follow the instructions below to import it manually:\n-----\n%s-----\n", c.Instructions()))
			return nil
		}
	}

	// get domain to put on the certificate name.
	// Linux does not have a proper certificate manager and we need to be able to identify the certificate
	domain, err := certDomain(file)
	if err != nil {
		return err
	}

	_, err = internal.RunCmd("sudo", "cp", file, fmt.Sprintf("/usr/local/share/ca-certificates/kyma-%s.crt", domain))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the Kyma certificates, please follow the instructions below to import them manually:\n-----\n%s-----\n", c.Instructions()))
	}
	_, err = internal.RunCmd("sudo", "update-ca-certificates")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("\nCould not import the Kyma certificates, please follow the instructions below to import them manually:\n-----\n%s-----\n", c.Instructions()))
	}

	return nil
}

func (certauth) Instructions() string {
	return "1. Download the certificate: kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\\.ingress\\.tlsCrt}' | base64 --decode > kyma.crt\n" +
		"2. Rename the certificate file: mv kyma.crt {NEW_CERT_NAME}\n" +
		"3. Copy the certificate to the CA folder: sudo cp {NEW_CERT_NAME} /usr/local/share/ca-certificates/\n" +
		"4. Update the certificate registry: sudo update-ca-certificates\n"
}

// certDomain returns the DNS info of the provided root certificate.
func certDomain(certFile string) (string, error) {
	certText, err := internal.RunCmd("openssl", "x509", "-text", "-noout", "-in", certFile)
	if err != nil {
		return "", err
	}

	matches := regexp.MustCompile("DNS:(.*)[\r\n]+").FindStringSubmatch(certText)

	if len(matches) < 2 {
		return "", errors.New("Could not determine the certificate's DNS")
	}

	return strings.Replace(matches[1], "'", "", -1), nil
}
