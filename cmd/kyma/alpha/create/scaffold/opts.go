package scaffold

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

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

func (o *Options) generateSecurityConfigFile() bool {
	return o.SecurityConfigFile != ""
}

func (o *Options) generateDefaultCRFile() bool {
	return o.DefaultCRFile != ""
}

var (
	fileNameModuleConfig = "module-config.yaml"
	//fileNameManifest        = "template-operator.yaml"
	fileNameSecurityConfig  = "sec-scanners-config.yaml"
	generatedDefaultCRFiles []string

	errInvalidDirectory             = errors.New("provided directory does not exist")
	errFilesExist                   = errors.New("scaffold already exists. use --overwrite flag to force scaffold creation")
	errModuleNameEmpty              = errors.New("--module-name flag must not be empty")
	errModuleVersionEmpty           = errors.New("--module-version flag must not be empty")
	errModuleChannelEmpty           = errors.New("--module-channel flag must not be empty")
	errManifestFileEmpty            = errors.New("--gen-manifest flag must not be empty")
	errModuleConfigEmpty            = errors.New("--module-config flag must not be empty")
	errInvalidManifestOptions       = errors.New("flag --gen-manifest cannot be set when argument --module-manifest-path provided")
	errManifestCreationFailed       = errors.New("could not generate manifest")
	errObjectsCreationFailed        = errors.New("could not generate webhook, rbac, and crd objects")
	errSecurityConfigCreationFailed = errors.New("could not generate security config")
	errDefaultCRCreationFailed      = errors.New("could not generate default CR")
	errModuleConfigCreationFailed   = errors.New("could not generate module config")
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
	/*
		err = o.validateFileOverwrite()
		if err != nil {
			return err
		}
		return o.validateManifestOptions()
	*/
	return nil
}

func (o *Options) validateDirectory() error {
	_, err := os.Stat(o.Directory)
	if errors.Is(err, os.ErrNotExist) {
		return errInvalidDirectory
	}
	absolutePath, err := filepath.Abs(o.Directory)
	if err != nil {
		return fmt.Errorf("error getting absolute file path to module directory: %w", err)
	}
	o.Directory = "/" + absolutePath
	return nil
}

/*
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
			entries, err := os.ReadDir(o.getCompleteFilePath(path.Join("config", "samples")))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("error while validating default CR directory: %w", err)
			}
			if len(entries) != 0 {
				return errFilesExist
			}
		}
	}
	return nil
}
*/

/*
func (o *Options) validateManifestOptions() error {
	if o.GenerateManifest && o.ModuleConfigManifestPath != "" {
		return errInvalidManifestOptions
	}
	return nil
}
*/

func (o *Options) getCompleteFilePath(fileName string) string {
	return path.Join(o.Directory, fileName)
}
