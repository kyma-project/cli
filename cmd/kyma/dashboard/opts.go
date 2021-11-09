package dashboard

import (
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the dashboard command
type Options struct {
	*cli.Options
	ContainerName string
	Port          string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

// validateFlags applies a sanity check on provided options
func (o *Options) validateFlags() error {
	if o.ContainerName == "" {
		return fmt.Errorf("either omit the --container-name flag or provide a value")
	}
	return nil
}
