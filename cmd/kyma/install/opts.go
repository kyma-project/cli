package install

import (
	"github.com/kyma-project/cli/internal/version"
	"github.com/pkg/errors"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	NoWait           bool
	Domain           string
	TLSCert          string
	TLSKey           string
	LocalSrcPath     string
	Timeout          time.Duration
	Password         string
	OverrideConfigs  []string
	ComponentsConfig string
	Source           string
	FallbackLevel    int
	CustomImage      string
	Profile          string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	// validate source flag
	kymaVersion, err := version.NewKymaVersion(o.Source)
	if err != nil {
		return errors.Errorf("Provided version (%s) is not a valid semantic version. It should be of format X.Y.Z", o.Source)
	}

	if kymaVersion.IsKyma2() {
		return errors.New("Kyma version 2.x can not be installed via 'install'. Please use the 'deploy' command, which supports Kyma 2 versions")
	}

	return nil
}
