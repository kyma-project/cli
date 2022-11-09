package module

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-project/cli/internal/cli"
)

// Options defines available options for the create module command
type Options struct {
	*cli.Options

	Name           string
	Version        string
	Path           string
	ModPath        string
	RegistryURL    string
	Credentials    string
	TemplateOutput string
	DefaultCRPath  string
	Channel        string
	Token          string
	Insecure       bool
	ResourcePaths  []string
	Overwrite      bool
	Clean          bool
}

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
