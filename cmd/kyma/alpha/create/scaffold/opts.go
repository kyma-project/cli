package scaffold

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
)

// Options specifies the flags for the scaffold command
type Options struct {
	*cli.Options

	Overwrite bool
	Directory string

	ModuleConfigFile   string
	ManifestFile       string
	SecurityConfigFile string
	DefaultCRFile      string

	ModuleName    string
	ModuleVersion string
	ModuleChannel string
}

func (o *Options) securityConfigFileConfigured() bool {
	return o.SecurityConfigFile != ""
}

func (o *Options) defaultCRFileConfigured() bool {
	return o.DefaultCRFile != ""
}

var (
	errDirNotExists                 = errors.New("provided directory does not exist")
	errNotDirectory                 = errors.New("provided path is not a directory")
	errModuleConfigExists           = errors.New("module config file already exists. use --overwrite flag to overwrite it")
	errModuleNameEmpty              = errors.New("--module-name flag must not be empty")
	errModuleVersionEmpty           = errors.New("--module-version flag must not be empty")
	errModuleChannelEmpty           = errors.New("--module-channel flag must not be empty")
	errManifestFileEmpty            = errors.New("--gen-manifest flag must not be empty")
	errModuleConfigEmpty            = errors.New("--module-config flag must not be empty")
	errManifestCreation             = errors.New("could not generate manifest")
	errDefaultCRCreationFailed      = errors.New("could not generate default CR")
	errModuleConfigCreationFailed   = errors.New("could not generate module config")
	errSecurityConfigCreationFailed = errors.New("could not generate security config")
)

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) Validate() error {
	if o.ModuleName == "" {
		return errModuleNameEmpty
	}

	if o.ModuleVersion == "" {
		return errModuleVersionEmpty
	}

	if o.ModuleChannel == "" {
		return errModuleChannelEmpty
	}

	err := o.validateDirectory()
	if err != nil {
		return err
	}

	if o.ModuleConfigFile == "" {
		return errModuleConfigEmpty
	}

	if o.ManifestFile == "" {
		return errManifestFileEmpty
	}

	return nil
}

func (o *Options) validateDirectory() error {
	fi, err := os.Stat(o.Directory)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", errDirNotExists, o.Directory)
		}
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("%w: %s", errNotDirectory, o.Directory)
	}

	return nil
}
