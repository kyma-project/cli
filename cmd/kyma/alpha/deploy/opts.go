package deploy

import (
	"errors"
	"fmt"
	"github.com/kyma-project/cli/internal/cli"
	"os"
)

const profileEvaluation = "evaluation"
const profileProduction = "production"

//Options defines available options for the command
type Options struct {
	*cli.Options

	Components     []string
	ComponentsFile string
	Domain         string
	Values         []string
	ValueFiles     []string
	Profile        string
	TLSCrtFile     string
	TLSKeyFile     string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	if err := o.validateProfile(); err != nil {
		return err
	}
	if err := o.validateTLSCertAndKey(); err != nil {
		return err
	}

	return nil
}

func (o *Options) validateProfile() error {
	if o.Profile == "" || o.Profile == profileEvaluation || o.Profile == profileProduction {
		return nil
	}

	return fmt.Errorf("unknown profile: %s", o.Profile)
}

func (o *Options) validateTLSCertAndKey() error {
	if o.TLSKeyFile == "" && o.TLSCrtFile == "" {
		return nil
	}
	if _, err := os.Stat(o.TLSKeyFile); os.IsNotExist(err) {
		return errors.New("tls key not found")
	}
	if _, err := os.Stat(o.TLSCrtFile); os.IsNotExist(err) {
		return errors.New("tls cert not found")
	}

	return nil
}
