package module

import (
	"github.com/kyma-project/cli/internal/cli"
)

// Options defines available options for the init module command
type Options struct {
	*cli.Options
	ModuleName string
	ParentDir  string
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
