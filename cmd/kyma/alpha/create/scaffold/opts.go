package scaffold

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
	"os"
	"path"
)

// Options specifies the flags for the scaffold command
type Options struct {
	*cli.Options

	Overwrite bool
	Directory string

	// Optional Files
	GenerateManifest       bool
	GenerateSecurityConfig bool
	GenerateDefaultCR      bool

	// Module Config Fields
	ModuleConfigName         string
	ModuleConfigVersion      string
	ModuleConfigChannel      string
	ModuleConfigManifestPath string
}

var (
	fileNameModuleConfig   = "module-config.yaml"
	fileNameManifest       = "template-operator.yaml"
	fileNameSecurityConfig = "sec-scanners-config.yaml"
	fileNameDefaultCR      = "config/samples/operator.kyma-project.io_v1alpha1_sample.yaml"

	errFilesExist                   = errors.New("scaffold already exists. use --overwrite flag to force scaffold creation")
	errInvalidManifestOptions       = errors.New("flag --gen-manifest cannot be set when argument --module-manifest-path provided")
	errManifestCreationFailed       = errors.New("could not generate manifest")
	errSecurityConfigCreationFailed = errors.New("could not generate security config")
	errDefaultCRCreationFailed      = errors.New("could not generate default CR")
	errModuleConfigCreationFailed   = errors.New("could not generate module config")
)

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) Validate() error {
	err := o.validateFileOverwrite()
	if err != nil {
		return err
	}
	return o.validateManifestOptions()
}

func (o *Options) validateFileOverwrite() error {
	if !o.Overwrite {
		_, err := os.Stat(o.getCompleteFilePath(fileNameModuleConfig))
		if !errors.Is(err, os.ErrNotExist) {
			return errFilesExist
		}

		if o.GenerateManifest {
			_, err := os.Stat(o.getCompleteFilePath(fileNameManifest))
			if !errors.Is(err, os.ErrNotExist) {
				return errFilesExist
			}
		}
		if o.GenerateSecurityConfig {
			_, err := os.Stat(o.getCompleteFilePath(fileNameSecurityConfig))
			if !errors.Is(err, os.ErrNotExist) {
				return errFilesExist
			}
		}
		if o.GenerateDefaultCR {
			_, err := os.Stat(o.getCompleteFilePath(fileNameDefaultCR))
			if !errors.Is(err, os.ErrNotExist) {
				return errFilesExist
			}
		}
	}
	return nil
}

func (o *Options) validateManifestOptions() error {
	if o.GenerateManifest && o.ModuleConfigManifestPath != "" {
		return errInvalidManifestOptions
	}
	return nil
}

func (o *Options) getCompleteFilePath(fileName string) string {
	dir := "./" + o.Directory
	return path.Join(dir, fileName)
}
