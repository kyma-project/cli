package mocks

import (
	"io/ioutil"

	"errors"

	"github.com/kyma-project/cli/internal/trust"
)

type Certifier struct {
	Crt string // mock certificate contents
}

func (c Certifier) Certificate() ([]byte, error) {
	if len(c.Crt) == 0 {
		// Mock not obtaining the certificate
		return nil, errors.New("Could not obtain certificate.")
	}
	return []byte(c.Crt), nil
}

func (c Certifier) StoreCertificate(file string, info trust.Informer) error {
	cert, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if string(cert) != c.Crt {
		return errors.New("Stored certificate not matching")
	}

	return nil
}

func (c Certifier) Instructions() string {
	return "Manual certificate import instructions OS specific."
}
