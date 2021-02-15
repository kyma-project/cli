package function

import (
	"os"
	"path"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
)

//Options defines available options for the command
type Options struct {
	*cli.Options

	Filename      string
	Dir           string
	ContainerName string
	FuncPort      string
	Detach        bool
	Debug         bool
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	options := &Options{Options: o}
	return options
}

func (o *Options) setDefaults() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if o.Filename == "" {
		o.Filename = path.Join(pwd, workspace.CfgFilename)
	}

	if o.Dir == "" {
		o.Dir = pwd
	}

	return nil
}
