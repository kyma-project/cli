package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-project/cli/internal/files"
	"github.com/pkg/errors"
)

func resolveLocalWorkspacePath(ws string, local bool) (string, error) {
	defaultWS := getDefaultWorkspacePath()

	if ws == "" {
		ws = defaultWS
	}
	//resolve local Kyma source directory only if user has not defined a custom workspace directory
	if local && ws == defaultWS {
		//use Kyma sources stored in GOPATH (if they exist)
		kymaRepo := filepath.Join("github.com", "kyma-project", "kyma")
		return resolveLocalRepo(kymaRepo)
	}

	if !local {
		if err := os.RemoveAll(ws); err != nil {
			return "", errors.Wrapf(err, "Could not delete old kyma source files in (%s)", ws)
		}
	}

	//no Kyma sources found in GOPATH
	return ws, nil
}

// resolveLocalRepo tries to find the repository with the given name in the GOPATH
// the repository must have its full name starting from $GOPATH/src
func resolveLocalRepo(repo string) (string, error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		goPath = filepath.Join(homeDir, "go")
	}
	goPath = filepath.Join(goPath, "src")

	path := filepath.Join(goPath, repo)
	if err := pathExists(path, "Local Kyma source directory"); err != nil {
		return "", err
	}

	return path, nil
}

func pathExists(path string, description string) error {
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
