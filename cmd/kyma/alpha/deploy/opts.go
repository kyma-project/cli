package deploy

import (
	"github.com/kyma-incubator/reconciler/pkg/reconciler/workspace"
	"github.com/kyma-project/cli/internal/cli"
	"path"
)

//Options defines available options for the command
type Options struct {
	*cli.Options
	Components     []string
	ComponentsFile string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

//ResolveComponentsFile makes overrides files locally available
func (o *Options) ResolveComponentsFile(workspace *workspace.Workspace) string {
	return path.Join(workspace.InstallationResourceDir, "components.yaml")
}
