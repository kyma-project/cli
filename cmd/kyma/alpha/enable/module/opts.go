package module

import (
	"fmt"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
)

const (
	CustomResourcePolicyCreateAndDelete = "CreateAndDelete"
	CustomResourcePolicyIgnore          = "Ignore"
)

type Options struct {
	*cli.Options

	Timeout   time.Duration
	Channel   string
	Namespace string
	KymaName  string
	Policy    string
	Force     bool
	Wait      bool
}

func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) validateFlags() error {
	if err := o.validateTimeout(); err != nil {
		return err
	}
	if err := o.validateChannel(); err != nil {
		return err
	}
	if err := o.validatePolicy(); err != nil {
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

func (o *Options) validateChannel() error {
	if o.Channel == "" {
		return nil
	}

	if len(o.Channel) < 3 {
		return errors.New("if provided, channel must be at least 3 chars long")
	}
	return nil
}

func (o *Options) validatePolicy() error {
	if o.Policy == CustomResourcePolicyCreateAndDelete || o.Policy == CustomResourcePolicyIgnore {
		return nil
	}

	return fmt.Errorf("policy must be either %s or %s", CustomResourcePolicyCreateAndDelete, CustomResourcePolicyIgnore)
}
