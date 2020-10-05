package function

import (
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"os"
)

//Options defines available options for the command
type Options struct {
	*cli.Options

	Name           string
	Namespace      string
	Dir            string
	Runtime        string
	URL            string
	RepositoryName string
	Reference      string
	BaseDir        string
	SourcePath     string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) setDefaults() (err error) {
	if o.Dir == "" {
		o.Dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	setIfZero(&o.SourcePath, o.Dir)
	return
}

func setIfZero(val *string, defaultValue string) {
	if *val == "" {
		*val = defaultValue
	}
}

func (o Options) source() workspace.Source {
	if o.URL != "" {
		return workspace.Source{
			SourceGit: workspace.SourceGit{
				BaseDir:    o.BaseDir,
				Reference:  o.Reference,
				Repository: o.RepositoryName,
				URL:        o.URL,
			},
			Type: workspace.SourceTypeGit,
		}
	}
	return workspace.Source{
		SourceInline: workspace.SourceInline{
			SourcePath: o.SourcePath,
		},
		Type: workspace.SourceTypeInline,
	}
}
