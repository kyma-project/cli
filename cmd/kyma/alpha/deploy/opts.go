package deploy

import (
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
)

// Options defines available options for the command
type Options struct {
	*cli.Options
	WildcardPermissions bool
	DryRun              bool
	OpenDashboard       bool
	CertManagerVersion  string
	ModulesFile         string
	Namespace           string
	Channel             string
	KymaCR              string
	Modules             []string
	Kustomizations      []string
	Templates           []string
	Timeout             time.Duration
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	if err := o.validateTimeout(); err != nil {
		return err
	}

	return nil
}

func (o *Options) validateTimeout() error {
	if o.Timeout <= 0 {
		return errors.New("timeout must be a positive duration")
	}
	return nil
}
