package trust

import (
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CertSourceConfigMap = "configmap"
	CertSourceSecret    = "secret"
)

type Source struct {
	name      string
	namespace string
	resource  string
}

func NewSource(name, namespace, resource string) Source {
	return Source{
		name:      name,
		namespace: namespace,
		resource:  resource,
	}
}

func certificateFromConfigMap(k keychain) ([]byte, error) {
	cm, err := k.k8s.Static().CoreV1().ConfigMaps(k.source.namespace).Get(k.source.name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	decodedCert, err := base64.StdEncoding.DecodeString(cm.Data["global.ingress.tlsCrt"])
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}

	return decodedCert, nil
}

func certificateFromSecret(k keychain) ([]byte, error) {
	secret, err := k.k8s.Static().CoreV1().Secrets(k.source.namespace).Get(k.source.name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("\nCould not retrieve the Kyma root certificate. Follow the instructions to import it manually:\n-----\n%s-----\n", k.Instructions()))
	}
	return secret.Data["tls.crt"], nil
}
