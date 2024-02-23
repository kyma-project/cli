package cli

import (
	"github.com/kyma-project/cli/pkg/step"
)

// Options defines available options for the command
type Options struct {
	CI      bool
	Verbose bool
	step.Factory
	KubeconfigPath string
	Finalizers     *Finalizers
}

// NewOptions creates options with default values
func NewOptions() *Options {
	return &Options{
		Finalizers: NewFinalizer(),
	}
}

// Synchronise the command options with the interactivity settings.
func (o *Options) Synchronise() {
	if o.CI {
		o.Factory.NonInteractive = true
	}

	if o.Verbose {
		o.Factory.UseLogger = true
	}
}
