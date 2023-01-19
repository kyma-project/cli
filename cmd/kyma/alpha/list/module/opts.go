package module

import (
	"fmt"
	"regexp"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
)

// Options defines available options for the create module command
type Options struct {
	*cli.Options

	Channel   string
	Timeout   time.Duration
	KymaName  string
	Namespace string
}

const (
	ChannelMinLength = 3
	ChannelMaxLength = 32
)

var (
	ErrChannelValidation = errors.New("channel validation failed")
)

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	if err := o.validateTimeout(); err != nil {
		return err
	}

	if err := o.validateChannel(); err != nil {
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
	if len(o.Channel) < ChannelMinLength || len(o.Channel) > ChannelMaxLength {
		return fmt.Errorf(
			"invalid channel length, length should between %d and %d, %w",
			ChannelMinLength, ChannelMaxLength, ErrChannelValidation,
		)
	}
	matched, _ := regexp.MatchString(`^[a-z]+$`, o.Channel)
	if !matched {
		return fmt.Errorf("invalid channel format, only allow characters from a-z")
	}
	return nil
}
