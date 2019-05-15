package uninstall

import (
	"github.com/kyma-project/cli/pkg/kyma/core"
	"time"
)

//Options defines available options for the command
type Options struct {
	*core.Options
	Timeout time.Duration
}

//NewOptions creates options with default values
func NewOptions(o *core.Options) *Options {
	return &Options{Options: o}
}

