package logs

import (
	"github.com/kyma-project/cli/internal/cli"
)

type Options struct {
	*cli.Options
	InStatus          string
	IngoredContainers []string
}

func NewOptions(o *cli.Options) *Options {
	return &Options{
		Options: o,
	}
}
