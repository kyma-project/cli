package version

import "github.com/kyma-project/cli/pkg/kyma/core"

//Options defines available options for the version command
type Options struct {
	*core.Options
	Client bool
}

//NewOptions creates options with default values
func NewOptions(o *core.Options) *Options {
	return &Options{Options: o}
}
