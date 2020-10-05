package function

import (
	"os"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	reflectcli "github.com/kyma-project/cli/pkg/reflect"
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

var (
	gitSourceOptionNames = []string{
		"RepositoryName",
		"Reference",
		"BaseDir",
	}
)

// IsValid checks if options are valid
func (o Options) IsValid() (err error) {
	if o.URL != "" {
		return nil
	}

	return reflectcli.NoneOf(o, gitSourceOptionNames)
}

func (o *Options) SetDefaults() (err error) {
	if o.Dir == "" {
		o.Dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	setIfZero(&o.RepositoryName, o.Name)
	setIfZero(&o.SourcePath, o.Dir)
	setIfZero(&o.BaseDir, "/")
	setIfZero(&o.Reference, "master")

	return
}

func setIfZero(val *string, defaultValue string) {
	if *val == "" {
		*val = defaultValue
	}
}

func (o Options) Source() workspace.Source {
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
