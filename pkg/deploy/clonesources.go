package deploy

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/pkg/git"
)

const (
	kymaURL = "https://github.com/kyma-project/kyma"
)

// CloneSources from Github
func CloneSources(stepFac StepFactory, workspacePath string, source string) error {
	if _, err := os.Stat(workspacePath); !os.IsNotExist(err) {
		question := stepFac.AddStep("Prepare Kyma download")
		if question.PromptYesNo(fmt.Sprintf("Workspace folder '%s' exists. Can it be deleted? ", workspacePath)) {
			if err := os.RemoveAll(workspacePath); err != nil {
				question.Failuref("Could not delete workspace folder")
				return err
			}
			question.Success()
		} else {
			question.Failure()
			return fmt.Errorf("Download stopped by user")
		}
	}

	downloadStep := stepFac.AddStep("Downloading Kyma into workspace folder")
	rev, err := git.ResolveRevision(kymaURL, source)
	if err != nil {
		return err
	}

	err = git.CloneRevision(kymaURL, workspacePath, rev)
	if err == nil {
		downloadStep.Success()
	} else {
		downloadStep.Failure()
	}

	return err
}
