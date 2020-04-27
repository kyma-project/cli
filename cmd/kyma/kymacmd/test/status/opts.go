package status

import (
	"github.com/kyma-project/cli/internal/cli"
)

type Options struct {
	*cli.Options
	Wait         bool
	OutputFormat string
}

func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
