package deploy

import (
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/kyma-project/cli/internal/version"
	"github.com/pkg/errors"
)

const (
	versionLocal      = "local"
	profileEvaluation = "evaluation"
	profileProduction = "production"
)

// Options defines available options for the command
type Options struct {
	*cli.Options
	values.Sources
	WorkspacePath  string
	Source         string
	Components     []string
	ComponentsFile string
	Profile        string
	WorkerPoolSize int
	Timeout        time.Duration
	DefaultWS      string
	DryRun         bool
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) IsLocal() bool {
	return o.Source == versionLocal
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	if err := o.validateProfile(); err != nil {
		return err
	}
	if err := o.validateSource(); err != nil {
		return err
	}
	if err := o.validateTLSCertAndKey(); err != nil {
		return err
	}
	return o.validateTimeout()
}

func (o *Options) validateSource() error {
	kymaVersion, err := version.NewKymaVersion(o.Source)
	if err != nil {
		return errors.Errorf("Provided version (%s) is not a valid semantic version. It must be of format X.Y.Z", o.Source)
	}

	if kymaVersion.IsKyma1() {
		return errors.New("Kyma version 1.x cannot be installed with 'deploy'. Use the 'install' command, which supports Kyma 1.x")
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

func (o *Options) validateTimeout() error {
	if o.Timeout <= 0 {
		return errors.New("timeout must be a positive duration")
	}
	return nil
}
