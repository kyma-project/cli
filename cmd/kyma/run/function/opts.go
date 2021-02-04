package function

import (
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"os"
	"path"
	"time"
)

//Options defines available options for the command
type Options struct {
	*cli.Options

	Filename      string
	ImageName     string
	ContainerName string
	FuncPort      string
	Envs          []string
	Timeout       time.Duration
	Detach        bool
	Debug         bool
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	options := &Options{Options: o}
	return options
}

func (o *Options) setDefaults() error {
	if o.Filename == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		o.Filename = path.Join(pwd, workspace.CfgFilename)
	}

	return nil
}
