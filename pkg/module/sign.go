package module

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha512"
	"github.com/pkg/errors"
)

type ComponentSignConfig struct {
	Name          string // Name of the module (mandatory)
	Version       string // Version of the module (mandatory)
	KeyPath       string // The private key used for signing (mandatory)
	SignatureName string // Name of the signature for signing
}

func Sign(cfg *ComponentSignConfig, remote *Remote) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	repo, err := remote.GetRepository()
	if err != nil {
		return err
	}

	cva, err := repo.LookupComponentVersion(cfg.Name, cfg.Version)
	if err != nil {
		return err
	}

	signReg := signing.DefaultRegistry()
	issuer := "kyma-project.ío/cli"

	key, err := RSAKey(cfg.KeyPath)
	if err != nil {
		return err
	}

	signReg.RegisterPrivateKey(cfg.SignatureName, key)

	return compdesc.Sign(
		ocm.DefaultContext().CredentialsContext(),
		cva.GetDescriptor(),
		key,
		signReg.GetSigner(rsa.Algorithm),
		signReg.GetHasher(sha512.Algorithm),
		cfg.SignatureName, issuer,
	)
}

func Verify(cfg *ComponentSignConfig, remote *Remote) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	repo, err := remote.GetRepository()
	if err != nil {
		return err
	}

	cva, err := repo.LookupComponentVersion(cfg.Name, cfg.Version)
	if err != nil {
		return err
	}

	signReg := signing.DefaultRegistry()

	key, err := RSAKey(cfg.KeyPath)
	if err != nil {
		return err
	}

	signReg.RegisterPublicKey(cfg.SignatureName, key)

	return compdesc.Verify(
		cva.GetDescriptor(),
		signReg, cfg.SignatureName,
	)
}

func (cfg *ComponentSignConfig) validate() error {
	if cfg.Name == "" {
		return errors.New("The module name cannot be empty")
	}
	if cfg.Version == "" {
		return errors.New("The module version cannot be empty")
	}
	if cfg.KeyPath == "" {
		return errors.New("The key path cannot be empty")
	}
	if cfg.SignatureName == "" {
		return errors.New("The signature name cannot be empty")
	}

	return nil
}

func RSAKey(pathToPrivateKey string) (interface{}, error) {
	privKeyFile, err := os.ReadFile(pathToPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to open private key file: %w", err)
	}

	block, _ := pem.Decode(privKeyFile)
	if block == nil {
		return nil, fmt.Errorf("unable to decode pem formatted block in key: %w", err)
	}
	untypedPrivateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	key, ok := untypedPrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("parsed private key is not of type *rsa.PrivateKey: %T", untypedPrivateKey)
	}
	return key, nil
}
