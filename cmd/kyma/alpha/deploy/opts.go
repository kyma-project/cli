package deploy

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/download"
	"github.com/kyma-project/cli/internal/files"
	"github.com/pkg/errors"
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	Components       []string
	ComponentsFile   string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
//ResolveComponentsFile makes overrides files locally available
func (o *Options) ResolveComponentsFile() (string, error) {
	kymaHome, err := files.KymaHome()
	if err != nil {
		return "", errors.Wrap(err, "Could not find or create Kyma home directory")
	}
	file, err := download.GetFile("https://raw.githubusercontent.com/kyma-project/kyma/main/installation/resources/components.yaml", kymaHome)
	return file, err
}