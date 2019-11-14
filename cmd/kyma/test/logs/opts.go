package logs

import (
	"github.com/kyma-project/cli/internal/cli"
)

type options struct {
	*cli.Options
	InStatus          string
	IngoredContainers []string
}

func NewOptions(o *cli.Options) *options {
	return &options{
		Options: o,
	}
}
