package uninstall

import (
	"fmt"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

const (
	quitTimeoutFactor = 1.25
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	WorkspacePath    string
	Timeout          time.Duration
	TimeoutComponent time.Duration
	Concurrency      int
	KeepCRDs         bool
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

//QuitTimeout returns the calculated duration of the installation quit timeout
func (o *Options) QuitTimeout() time.Duration {
	return time.Duration((o.Timeout.Seconds() * quitTimeoutFactor)) * time.Second
}

// validateFlags applies a sanity check on provided options
func (o *Options) validateFlags() error {
	if o.Timeout < o.TimeoutComponent {
		return fmt.Errorf("Timeout (%v) cannot be smaller than component timeout (%v)", o.Timeout, o.TimeoutComponent)
	}
	return nil
}
