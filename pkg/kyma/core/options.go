package core

import (
	"github.com/kyma-incubator/kymactl/internal/kube"
	"github.com/kyma-incubator/kymactl/internal/step"
)

//Options defines available options for the command
type Options struct {
	Verbose bool
	step.Factory
	kube.ConfigFactory
}

//NewOptions creates options with default values
func NewOptions() *Options {
	return &Options{}
}
