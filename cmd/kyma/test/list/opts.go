package list

import "github.com/kyma-project/cli/internal/cli"

type options struct {
	*cli.Options
}

func NewOptions(o *cli.Options) *options {
	return &options{Options: o}
}
