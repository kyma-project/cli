package k3d

import (
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

// Options defines available options for the k3d provisioning command
type Options struct {
	*cli.Options

	Name              string
	Workers           int
	Timeout           time.Duration
	K3sArgs           []string
	UseRegistry       []string
	RegistryPort      string
	K3dArgs           []string
	K3dRegistryArgs   []string
	KubernetesVersion string
	PortMapping       []string
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
