package run

import (
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

type options struct {
	*cli.Options
	Name           string
	Watch          bool
	Timeout        time.Duration
	ExecutionCount int64
	MaxRetries     int64
	Concurrency    int64
}

func NewOptions(o *cli.Options) *options {
	return &options{Options: o}
}
