package run

import (
	"github.com/kyma-project/cli/pkg/kyma/core"
)

type options struct {
	*core.Options
	Name    string
	Tests   string
	Wait    bool
	Timeout int
}

func NewOptions(o *core.Options) *options {
	return &options{Options: o}
}
