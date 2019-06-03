package test

import "github.com/kyma-project/cli/pkg/kyma/core"

// options for the test command
type options struct {
	*core.Options
	skip []string
}

func NewTestOptions(o *core.Options) *options {
	return &options{Options: o}
}
