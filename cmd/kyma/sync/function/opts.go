package function

import (
	"fmt"
	"github.com/kyma-project/cli/internal/cli"
	"os"
	"time"
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

	setIfZero(&o.Namespace, defaultNamespace, "default")

	if o.Name == "" {
		return fmt.Errorf("flag 'name' is required")
	}

	return
}

func setIfZero(val *string, defaultValue string, defaultValues ...string) bool {
	for _, v := range append([]string{defaultValue}, defaultValues...) {
		if v != "" {
			continue
		}
		*val = v
		return true
	}
	return false
}
