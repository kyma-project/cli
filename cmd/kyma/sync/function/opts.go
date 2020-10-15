package function

import (
	"github.com/kyma-project/cli/internal/cli"
	"os"
)

//Options defines available options for the command
type Options struct {
	*cli.Options

	Name           string
	Namespace      string
	OutputPath     string
	Kubeconfig     string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	options := &Options{Options: o}
	return options
}

func (o *Options) setDefaults() (err error) {
	if o.OutputPath == "" {
		o.OutputPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	setIfZero(&o.Namespace, defaultNamespace)
	return
}

func setIfZero(val *string, defaultValue string) {
	if *val == "" {
		*val = defaultValue
	}
}
