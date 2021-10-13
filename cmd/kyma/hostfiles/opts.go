package hostfiles

import (
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	Domain string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) validateFlags() error {
	if o.Domain == "" {
		return fmt.Errorf("--add the domain name")
	}
	return nil
}
