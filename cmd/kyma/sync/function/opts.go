package function

import (
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the command
type Options struct {
	*cli.Options

	Name       string
	Namespace  string
	OutputPath string
	Timeout    time.Duration
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	options := &Options{Options: o}
	return options
}

func (o *Options) setDefaults(defaultNamespace string) (err error) {
	if o.OutputPath == "" {
		o.OutputPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	if o.Namespace == "" {
		o.Namespace = defaultNamespace
	}

	if o.Name == "" {
		return fmt.Errorf("flag 'name' is required")
	}

	return
}
