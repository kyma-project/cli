// +build windows

package trust

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/root"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type certutil struct {
	k8s kube.KymaKube
}

func NewCertifier(k kube.KymaKube) Certifier {
	return certutil{
		k8s: k,
	}
}

func (c certutil) Certificate() ([]byte, error) {
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

func (c certutil) CertificateAlpha() ([]byte, error) {
	s, err := c.k8s.Static().CoreV1().Secrets("istio-system").Get(context.Background(), "kyma-gateway-certs", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", c.Instructions()))
	}

	return s.Data["tls.crt"], nil
}

func (c certutil) StoreCertificate(file string, i Informer) error {
	i.LogInfo("Kyma wants to add its root certificate to the trusted certificates.")

	if root.IsWithSudo() {
		i.LogInfo("You're running CLI with sudo. CLI has to add the Kyma root certificate to the trusted certificates. Type 'y' to allow this action.")
		if !root.PromptUser() {
			i.LogInfo(fmt.Sprintf("\nCould not import the Kyma root certificate, please follow the instructions below to import it manually:\n-----\n%s-----\n", c.Instructions()))
			return nil
		}
		// Only automatically add the cert if already on admin mode, can't ask for admin password from go
		_, err := cli.RunCmd("certutil", "-addstore", "-f", "Root", file)
		return err
	}
	return errors.New(fmt.Sprintf("Could not import the Kyma root certificate. Follow the instructions to import them manually:\n-----\n%s-----\n", c.Instructions()))
}

func (certutil) Instructions() string {
	return "1. Open a terminal window with administrator rights.\n" +
		"2. Download the certificate: kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\\.ingress\\.tlsCrt}' > tmp.txt\n" +
		"3. Decode the certificate: certutil -decode tmp.txt kyma.crt ; del tmp.txt\n" +
		"4. Import the certificate: certutil -addstore -f Root kyma.crt\n"
}

func (certutil) InstructionsAlpha() string {
	return "1. Open a terminal window with administrator rights.\n" +
		"2. Download the certificate: kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\\.crt}' > tmp.txt\n" +
		"3. Decode the certificate: certutil -decode tmp.txt kyma.crt ; del tmp.txt\n" +
		"4. Import the certificate: certutil -addstore -f Root kyma.crt\n"
}
