package module

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Options defines available options for the create module command
type Options struct {
	*cli.Options

	KymaName      string
	Channel       string
	Timeout       time.Duration
	Namespace     string
	AllNamespaces bool
	NoHeaders     bool
	Output        string
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
	if err := o.validateOutput(); err != nil {
		return err
	}
	if err := o.validateTimeout(); err != nil {
		return err
	}
	if err := o.validateChannel(); err != nil {
		return err
	}

	if o.AllNamespaces {
		o.Namespace = metav1.NamespaceAll
	}

	return nil
}

func (o *Options) validateOutput() error {
	valids := []string{
		"json",
		"yaml",
		"go-template-file",
	}
	for _, valid := range valids {
		if o.Output == valid {
			return nil
		}
	}
	return fmt.Errorf("output must be one of: (%s)", strings.Join(valids, ", "))
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
	matched, err := regexp.MatchString(`^[a-z]+$`, o.Channel)
	if err != nil {
		return nil
	}
	if !matched {
		return fmt.Errorf("invalid channel format, only allow characters from a-z")
	}
	return nil
}
