package function

import (
	"os"

	"github.com/kyma-incubator/hydroform/function/pkg/generator"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
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
	VsCode         bool
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	options := &Options{Options: o}
	return options
}

func (o *Options) setDefaults(defaultNamespace string) (err error) {
	if o.Dir == "" {
		o.Dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	if o.Name == "" {
		generated, err := generator.GenerateName(true)
		if err != nil {
			return err
		}
		o.Name = "function-" + generated
	}

	if o.RepositoryName == "" {
		o.RepositoryName = o.Name
	}

	setIfZero(&o.Namespace, defaultNamespace)
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
