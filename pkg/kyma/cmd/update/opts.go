package update

import (
	"time"

	"github.com/kyma-incubator/kyma-cli/pkg/kyma/core"
)

//Options defines available options for the command
type Options struct {
	*core.Options
	ReleaseVersion string
	NoWait         bool
	Timeout        time.Duration
}

//NewOptions creates options with default values
func NewOptions(o *core.Options) *Options {
	return &Options{Options: o}
}
