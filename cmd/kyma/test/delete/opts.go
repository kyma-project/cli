package del

import "github.com/kyma-project/cli/internal/cli"

type Options struct {
	*cli.Options
	Name string
	All  bool
}

func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
