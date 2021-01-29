package deploy

import (
	"os"

	"github.com/kyma-project/cli/pkg/git"
	"github.com/kyma-project/cli/pkg/step"
)

const (
	kymaURL = "https://github.com/kyma-project/kyma"
)

// CloneSources from Github
func CloneSources(stepFac step.FactoryInterface, workspacePath string, source string) error {
	if _, err := os.Stat(workspacePath); !os.IsNotExist(err) {
		prepDownloadStep := stepFac.NewStep("Prepare Kyma download")
		if err := os.RemoveAll(workspacePath); err != nil {
			prepDownloadStep.Failuref("Could not delete workspace folder")
			return err
		}
		prepDownloadStep.Success()
	}

	downloadStep := stepFac.NewStep("Downloading Kyma into workspace folder")
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
