package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"regexp"

	"github.com/blang/semver/v4"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
)

// Options defines available options for the create module command
type Options struct {
	*cli.Options

	Name                    string
	NameMappingMode         string
	Version                 string
	Path                    string
	ModuleArchivePath       string
	RegistryURL             string
	Credentials             string
	TemplateOutput          string
	DefaultCRPath           string
	Channel                 string
	Target                  string
	SchemaVersion           string
	Token                   string
	Insecure                bool
	PersistentArchive       bool
	ResourcePaths           []string
	ArchiveVersionOverwrite bool
	RegistryCredSelector    string
	SecurityScanConfig      string
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

func (o *Options) ValidateVersion() error {
	sv, err := semver.ParseTolerant(o.Version)
	if err != nil {
		return err
	}
	o.Version = sv.String()
	if !strings.HasPrefix(o.Version, "v") {
		o.Version = fmt.Sprintf("v%s", o.Version)
	}
	return nil
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

func (o *Options) ValidateTarget() error {
	valid := []string{
		"control-plane",
		"remote",
	}
	for i := range valid {
		if o.Target == valid[i] {
			return nil
		}
	}
	return fmt.Errorf("target %s is invalid, allowed: %s", o.Target, valid)
}
