package function

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/function/pkg/generator"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"math/rand"
	"os"
	"time"
)

//Options defines available options for the command
type Options struct {
	*cli.Options

	Name           string
	Namespace      string
	Dir            string
	Runtime        string
	URL            string
	RepositoryName repositoryName
	Reference      string
	BaseDir        string
	SourcePath     string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	options := &Options{Options: o}
	options.RepositoryName = newRepositoryName(&options.Name)
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
		rand.Seed(time.Now().UnixNano())
		o.Name = "function-" + generator.GenerateName(true)
	}

	setIfZero(&o.SourcePath, o.Dir)
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
				Repository: o.RepositoryName.String(),
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

type repositoryName struct {
	value *string
}

func newRepositoryName(defaultVal *string) repositoryName {
	return repositoryName{value: defaultVal}
}

func (rn repositoryName) String() string {
	if rn.value != nil {
		return *rn.value
	}
	return ""
}

func (rn *repositoryName) Set(v string) error {
	if rn == nil {
		return fmt.Errorf("nil pointer reference")
	}
	if v != "" {
		rn.value = &v
	}

	return nil
}

func (rn repositoryName) Type() string {
	return "string"
}
