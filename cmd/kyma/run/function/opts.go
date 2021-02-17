package function

import (
	"os"
	"path"
	"path/filepath"

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

func (o *Options) defaultFilename() error {
	if o.Filename == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		o.Filename = path.Join(pwd, workspace.CfgFilename)
	} else if !filepath.IsAbs(o.Filename) {
		filename, err := filepath.Abs(o.Filename)
		if err != nil {
			return err
		}
		o.Filename = filename
	}

	return nil
}

func (o *Options) defaultValues(cfg workspace.Cfg) {
	if o.Dir == "" && cfg.Source.Type == workspace.SourceTypeInline {
		o.Dir = cfg.Source.SourcePath
	} else if o.Dir == "" {
		o.Dir = filepath.Dir(o.Filename)
	}

	if o.ContainerName == "" {
		o.ContainerName = cfg.Name
	}
}
