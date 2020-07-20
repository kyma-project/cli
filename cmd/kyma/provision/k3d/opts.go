package k3d

import (
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

//options defines available options for the k3d provisioning command
type Options struct {
	*cli.Options

	PublishHTTP    string
	PublishHTTPS   string
	EnableRegistry string
	RegistryVolume string
	RegistryName   string
	ServerArg      string
	NoDeploy       string
	Name           string
	Timeout        time.Duration
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
