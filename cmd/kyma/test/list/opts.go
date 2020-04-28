package list

import "github.com/kyma-project/cli/internal/cli"

type Options struct {
	*cli.Options
}

func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
