package status

import (
	"github.com/kyma-project/cli/internal/cli"
)

type options struct {
	*cli.Options
	Wait         bool
	Logs         string
	OutputFormat string
}

func NewOptions(o *cli.Options) *options {
	return &options{Options: o}
}
