package run

import (
	"time"

	"github.com/kyma-project/cli/pkg/kyma/core"
)

type options struct {
	*core.Options
	Name           string
	Wait           bool
	Timeout        time.Duration
	ExecutionCount int64
	MaxRetries     int64
	Concurrency    int64
}

func NewOptions(o *core.Options) *options {
	return &options{Options: o}
}
