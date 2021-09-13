package deploy

import (
	"fmt"
	"github.com/kyma-project/cli/internal/cli"
)

const profileEvaluation = "evaluation"
const profileProduction = "production"

//Options defines available options for the command
type Options struct {
	*cli.Options

	Components     []string
	ComponentsFile string
	Domain         string
	Values         []string
	ValueFiles []string
	Profile        string
	TLSCrtFile     string
	TLSKeyFile     string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

// validateFlags performs a sanity check of provided options
func (o *Options) validateFlags() error {
	return o.validateProfile()
}

func (o *Options) validateProfile() error {
	switch o.Profile {
	case "":
	case profileEvaluation:
	case profileProduction:
		return nil
	}

	return fmt.Errorf("unknown profile: %s", o.Profile)
}
