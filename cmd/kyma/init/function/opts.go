package function

import "github.com/kyma-project/cli/internal/cli"

//Options defines available options for the command
type Options struct {
	*cli.Options

	Name      string
	Namespace string
	Dir       string
	Runtime   string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
