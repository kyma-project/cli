package list

import "github.com/kyma-project/cli/pkg/kyma/core"

type options struct {
	*core.Options
}

func NewOptions(o *core.Options) *options {
	return &options{Options: o}
}
