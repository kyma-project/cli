package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"

	"github.com/kyma-incubator/reconciler/pkg/reconciler/chart"
	"github.com/kyma-project/cli/internal/files"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func isWorkspaceEmpty(workspace string) (bool, error) {
	_, err := os.Stat(workspace)
	if !os.IsNotExist(err) {
		isWorkspaceEmpty, err := files.IsDirEmpty(workspace)
		if err != nil {
			return false, err
		}
		if !isWorkspaceEmpty && workspace != getDefaultWorkspacePath() {
			return false, nil
		}
	}
	return true, nil
}

func PrepareDryRunWorkspace(workspace, source string, local bool, l *zap.SugaredLogger) (*chart.KymaWorkspace, error) {
	if !local {
		isEmpty, err := isWorkspaceEmpty(workspace)
		if err != nil {
			return nil, err
		}
		if !isEmpty {
			return nil, errors.Errorf("Workspace '%s' not empty. Aborting deployment", workspace)
		}
	}
	return fetchWorkspace(workspace, source, local, l)
}

func fetchWorkspace(workspace, source string, local bool, l *zap.SugaredLogger) (*chart.KymaWorkspace, error) {
	wsp, err := resolveLocalWorkspacePath(workspace, local)
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

	return ws, nil
}

func PrepareWorkspace(workspace, source string, step step.Step, interactive, local bool, l *zap.SugaredLogger) (*chart.KymaWorkspace, error) {
	if !local && interactive {
		isEmpty, err := isWorkspaceEmpty(workspace)
		if err != nil {
			return nil, err
		}
		if !isEmpty && !step.PromptYesNo(fmt.Sprintf("Existing files in workspace folder '%s' will be deleted. Are you sure you want to continue? ", workspace)) {
			step.Failure()
			return nil, errors.Errorf("aborting deployment")
		}
	}

	ws, err := fetchWorkspace(workspace, source, local, l)
	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch workspace")
	}

	step.Successf("Using Kyma from the workspace directory: %s", ws.WorkspaceDir)

	return ws, nil
}

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
