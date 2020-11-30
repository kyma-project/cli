package install

import (
	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	OverridesYaml  string
	ComponentsYaml string
	ResourcesPath  string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
