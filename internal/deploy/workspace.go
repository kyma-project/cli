package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-incubator/reconciler/pkg/reconciler/chart"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func PrepareWorkspace(workspace, source string, step step.Step, interactive, local bool, l *zap.SugaredLogger) (*chart.KymaWorkspace, error) {
	if !local {
		_, err := os.Stat(workspace)
		if !os.IsNotExist(err) && interactive {
			isWorkspaceEmpty, err := files.IsDirEmpty(workspace)
			if err != nil {
				return nil, err
			}
			if !isWorkspaceEmpty && workspace != getDefaultWorkspacePath() {
				if !step.PromptYesNo(fmt.Sprintf("Existing files in workspace folder '%s' will be deleted. Are you sure you want to continue? ", workspace)) {
					step.Failure()
					return nil, errors.Errorf("aborting deployment")
				}
			}
		}
	}

	wsp, err := ResolveLocalWorkspacePath(workspace, local)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve workspace path")
	}

	wsFact, err := chart.NewFactory(nil, wsp, l)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate workspace factory")
	}

	err = service.UseGlobalWorkspaceFactory(wsFact)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set global workspace factory")
	}

	ws, err := wsFact.Get(source)
	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch workspace")
	}

	step.Successf("Using Kyma from the workspace directory: %s", wsp)

	return ws, nil
}

func ResolveLocalWorkspacePath(ws string, local bool) (string, error) {
	defaultWS := getDefaultWorkspacePath()

	if ws == "" {
		ws = defaultWS
	}
	//resolve local Kyma source directory only if user has not defined a custom workspace directory
	if local && ws == defaultWS {
		//use Kyma sources stored in GOPATH (if they exist)
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			goPath = filepath.Join(homeDir, "go")
		}
		kymaPath := filepath.Join(goPath, "src", "github.com", "kyma-project", "kyma")
		if pathExists(kymaPath, "Local Kyma source directory") == nil {
			return kymaPath, nil
		}
	}

	if !local {
		if err := os.RemoveAll(ws); err != nil {
			return "", errors.Wrapf(err, "Could not delete old kyma source files in (%s)", ws)
		}
	}

	//no Kyma sources found in GOPATH
	return ws, nil
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
