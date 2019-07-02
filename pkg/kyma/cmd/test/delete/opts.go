package del

import "github.com/kyma-project/cli/pkg/kyma/core"

type options struct {
	*core.Options
	Name string
	All  bool
}

func NewOptions(o *core.Options) *options {
	return &options{Options: o}
}
