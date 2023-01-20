package module

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
	"regexp"
)

// Options defines available options for the create module command
type Options struct {
	*cli.Options

	Name                 string
	Version              string
	Path                 string
	ModCache             string
	RegistryURL          string
	Credentials          string
	TemplateOutput       string
	DefaultCRPath        string
	Channel              string
	Token                string
	Insecure             bool
	ResourcePaths        []string
	Overwrite            bool
	Clean                bool
	RegistryCredSelector string
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

func (o *Options) ValidatePath() error {
	var err error
	if o.Path == "" {
		o.Path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("could not ge the current directory: %w", err)
		}
	} else {
		o.Path, err = filepath.Abs(o.Path)
		if err != nil {
			return fmt.Errorf("could not obtain absolute path to module %q: %w", o.Path, err)
		}
	}
	return err
}

func (o *Options) ValidateChannel() error {

	if len(o.Channel) < ChannelMinLength || len(o.Channel) > ChannelMaxLength {
		return fmt.Errorf("invalid channel length, length should between %d and %d, %w",
			ChannelMinLength, ChannelMaxLength, ErrChannelValidation)
	}
	matched, _ := regexp.MatchString(`^[a-z]+$`, o.Channel)
	if !matched {
		return fmt.Errorf("invalid channel format, only allow characters from a-z")
	}
	return nil
}
