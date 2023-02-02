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
}

func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

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
