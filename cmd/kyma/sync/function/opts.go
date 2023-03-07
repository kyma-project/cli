package function

import (
	"os"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

// Options defines available options for the command
type Options struct {
	*cli.Options

	Namespace     string
	Dir           string
	Timeout       time.Duration
	SchemaVersion string
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	options := &Options{Options: o}
	return options
}

func (o *Options) setDefaults(defaultNamespace string) (err error) {
	if o.Dir == "" {
		o.Dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	if o.Namespace == "" {
		o.Namespace = defaultNamespace
	}

	return
}
