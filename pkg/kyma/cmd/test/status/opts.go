package status

import (
	"github.com/kyma-project/cli/pkg/kyma/core"
)

type options struct {
	*core.Options
	Wait bool
	Jsn  bool
	Logs string
}

func NewOptions(o *core.Options) *options {
	return &options{Options: o}
}
