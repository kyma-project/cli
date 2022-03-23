package minikube

import (
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

//options defines available options for the minikube provisioning command
type Options struct {
	*cli.Options
	VMDriver            string
	DiskSize            string
	Memory              string
	CPUS                string
	HypervVirtualSwitch string
	DockerPorts         []string
	Profile             string
	UseVPNKitSock       bool
	Timeout             time.Duration
	KubernetesVersion   string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
