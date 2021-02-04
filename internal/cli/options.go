package cli

import (
	"github.com/kyma-project/cli/pkg/step"
)

//Options defines available options for the command
type Options struct {
	CI      bool
	Verbose bool
	step.Factory
	KubeconfigPath string
	Finalizer      *Finalizer
}

//NewOptions creates options with default values
func NewOptions(finalizer *Finalizer) *Options {
	return &Options{
		Finalizer: finalizer,
	}
}
