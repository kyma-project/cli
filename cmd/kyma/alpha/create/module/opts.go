package module

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kyma-project/cli/internal/nice"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/cli"
)

// Options defines available options for the create module command
type Options struct {
	*cli.Options

	Name                    string
	NameMappingMode         string
	Version                 string
	Path                    string
	ModuleArchivePath       string
	GitRemote               string
	RegistryURL             string
	Credentials             string
	TemplateOutput          string
	DefaultCRPath           string
	Channel                 string
	SchemaVersion           string
	Token                   string
	Insecure                bool
	PersistentArchive       bool
	ResourcePaths           []string
	ArchiveVersionOverwrite bool
	RegistryCredSelector    string
	SecurityScanConfig      string
	PrivateKeyPath          string
	ModuleConfigFile        string
	KubebuilderProject      bool
	Namespace               string
	DryRun                  bool
}

const (
	ChannelMinLength = 3
	ChannelMaxLength = 32
)

var (
	ErrChannelValidation       = errors.New("channel validation failed")
	ErrManifestFetch           = errors.New("remote YAML manifest fetch failed")
	ErrManifestPathValidation  = errors.New("YAML manifest path validation failed")
	ErrDefaultCRFetch          = errors.New("remote default CR fetch failed")
	ErrDefaultCRPathValidation = errors.New("default CR path validation failed")
	ErrNameValidation          = errors.New("name validation failed")
	ErrNamespaceValidation     = errors.New("namespace validation failed")
	ErrVersionValidation       = errors.New("version validation failed")
)

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) validateVersion() error {
	sv, err := semver.NewVersion(o.Version)
	if err != nil {
		return err
	}
	o.Version = sv.String()
	return nil
}

func (o *Options) validatePath() error {
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

func (o *Options) validateChannel() error {
	if len(o.Channel) < ChannelMinLength || len(o.Channel) > ChannelMaxLength {
		return fmt.Errorf(
			"invalid channel length, length should between %d and %d, %w",
			ChannelMinLength, ChannelMaxLength, ErrChannelValidation,
		)
	}
	matched, err := regexp.MatchString(`^[a-z]+$`, o.Channel)
	if err != nil {
		return fmt.Errorf("failed to evaluate regex for channel: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid channel format, only allow characters from a-z")
	}
	return nil
}

func (o *Options) Validate() error {
	if o.KubebuilderProject {
		if err := o.validateVersion(); err != nil {
			return err
		}

		if err := o.validateChannel(); err != nil {
			return err
		}
	} else if !o.WithModuleConfigFile() {
		np := nice.Nice{}
		np.PrintImportant("WARNING: \"--module-config-file\" flag is required. If you want to build a module " +
			"from a Kubebuilder project instead, use the \"--kubebuilder-project\" flag.")
		err := errors.New("\"--module-config-file\" flag is required")
		return err
	}

	return o.validatePath()
}

func (o *Options) WithModuleConfigFile() bool {
	return len(o.ModuleConfigFile) > 0
}
