package module

import (
	"github.com/kyma-project/cli/internal/cli"
)

// Options defines available options for the version command
type Options struct {
	*cli.Options
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
