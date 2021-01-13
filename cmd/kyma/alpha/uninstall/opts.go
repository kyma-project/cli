package uninstall

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

var (
	defaultWorkspacePath = filepath.Join(".", "workspace")
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	WorkspacePath string
	OverridesFile string
	CancelTimeout time.Duration
	QuitTimeout   time.Duration
	HelmTimeout   time.Duration
	WorkersCount  int
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

//getDefaultWorkspacePath return the default path to the CLI workspace directory
func (o *Options) defaultWorkspacePath() string {
	return defaultWorkspacePath
}

// validateFlags applies a sanity check on provided options
func (o *Options) validateFlags() error {
	if o.QuitTimeout < o.CancelTimeout {
		return fmt.Errorf("Quit timeout (%v) cannot be smaller than cancel timeout (%v)", o.QuitTimeout, o.CancelTimeout)
	}
	return nil
}
