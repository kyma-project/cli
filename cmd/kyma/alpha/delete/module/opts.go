package module

import (
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
)

type Options struct {
	*cli.Options

	Timeout   time.Duration
	Channel   string
	Namespace string
	KymaName  string
	Force     bool
	Wait      bool
}

func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) validateFlags() error {
	return o.validateTimeout()
}

func (o *Options) validateTimeout() error {
	if o.Timeout <= 0 {
		return errors.New("timeout must be a positive duration")
	}
	return nil
}
