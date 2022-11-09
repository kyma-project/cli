package module

import (
	"github.com/kyma-project/cli/internal/cli"
)

// Options defines available options for the create module command
type Options struct {
	*cli.Options

	ModPath        string
	RegistryURL    string
	Credentials    string
	TemplateOutput string
	Channel        string
	Token          string
	Insecure       string
	ResourcePaths  []string
	Overwrite      bool
	Clean          bool
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
