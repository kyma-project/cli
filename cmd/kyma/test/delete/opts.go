package del

import "github.com/kyma-project/cli/internal/cli"

type options struct {
	*cli.Options
	Name string
	All  bool
}

func NewOptions(o *cli.Options) *options {
	return &options{Options: o}
}
