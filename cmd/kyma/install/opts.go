package install

import (
	"time"

	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	NoWait           bool
	Domain           string
	TLSCert          string
	TLSKey           string
	LocalSrcPath     string
	Timeout          time.Duration
	Password         string
	OverrideConfigs  []string
	ComponentsConfig string
	Source           string
	FallbackLevel    int
	CustomImage      string
	Profile          string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
