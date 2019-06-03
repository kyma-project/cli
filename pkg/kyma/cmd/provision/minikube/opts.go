package minikube

import "github.com/kyma-project/cli/pkg/kyma/core"

//options defines available options for the minikube provisioning command
type options struct {
	*core.Options
	Domain              string
	VMDriver            string
	DiskSize            string
	Memory              string
	CPU                 string
	HypervVirtualSwitch string
}

//NewOptions creates options with default values
func NewOptions(o *core.Options) *options {
	return &options{Options: o}
}
