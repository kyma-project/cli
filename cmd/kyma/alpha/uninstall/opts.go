package uninstall

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

var (
	defaultWorkspacePath  = filepath.Join(".", "workspace")
	defaultResourcePath   = filepath.Join(defaultWorkspacePath, "kyma", "resources")
	defaultComponentsFile = filepath.Join(defaultWorkspacePath, "kyma", "installation", "resources", "components.yaml")
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	WorkspacePath      string
	ResourcePath       string
	OverridesFile      string
	ComponentsListFile string
	CancelTimeout      time.Duration
	QuitTimeout        time.Duration
	HelmTimeout        time.Duration
	WorkersCount       int
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

//getDefaultResourcesPath return the default path to the Kyma resources directory
func (o *Options) getDefaultResourcePath() string {
	return defaultResourcePath
}

//getDefaultWorkspacePath return the default path to the CLI workspace directory
func (o *Options) getDefaultWorkspacePath() string {
	return defaultWorkspacePath
}

//getDefaultComponentsFile return the default path to the Kyma components file
func (o *Options) getDefaultComponentsListFile() string {
	return defaultComponentsFile
}

// validateFlags applies a sanity check on provided options
func (o *Options) validateFlags() error {
	if err := o.pathExists(o.ResourcePath, "Resource path"); err != nil {
		return err
	}
	if o.QuitTimeout < o.CancelTimeout {
		return fmt.Errorf("Quit timeout (%v) cannot be smaller than cancel timeout (%v)", o.QuitTimeout, o.CancelTimeout)
	}
	return nil
}

func (o *Options) pathExists(path string, description string) error {
	if path == "" {
		return fmt.Errorf("%s is empty", description)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s '%s' not found", description, path)
	}
	return nil
}
