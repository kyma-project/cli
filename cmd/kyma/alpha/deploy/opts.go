package deploy

import (
	"fmt"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
	"os"
	"path/filepath"
)

const VersionLocal = "local"
var (
defaultWorkspacePath  = getDefaultWorkspacePath()
defaultComponentsFile = filepath.Join(defaultWorkspacePath, "installation", "resources", "components.yaml")
)
//Options defines available options for the command
type Options struct {
	*cli.Options
	WorkspacePath    string
	Source           string

}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}

func (o *Options) ResolveLocalWorkspacePath() string {
	if o.WorkspacePath == "" {
		o.WorkspacePath = defaultWorkspacePath
	}
	//resolve local Kyma source directory only if user has not defined a custom workspace directory
	if o.Source == VersionLocal && o.WorkspacePath == defaultWorkspacePath {
		//use Kyma sources stored in GOPATH (if they exist)
		goPath := os.Getenv("GOPATH")
		if goPath != "" {
			kymaPath := filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
			if o.pathExists(kymaPath, "Local Kyma source directory") == nil {
				return kymaPath
			}
		}
	}
	//no Kyma sources found in GOPATH
	return o.WorkspacePath
}

func (o *Options) pathExists(path string, description string) error {
	if path == "" {
		return fmt.Errorf("%s is empty", description)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s '%s' not found", description, path)
	}
	return nil
}

func getDefaultWorkspacePath() string {
	kymaHome, err := files.KymaHome()
	if err != nil {
		return ".kyma-sources"
	}
	return filepath.Join(kymaHome, "sources")
}