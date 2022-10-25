package cli

import (
	"errors"

	"github.com/kyma-project/cli/internal/clusterinfo"
)

const managedKymaWarning = "CAUTION: You are trying to use Kyma CLI to change a managed Kyma runtime (SAP Kyma Runtime). This action may corrupt the Kyma runtime. Proceed at your own risk."

// DetectManagedEnvironment introduces a step that checks if the target runtime is a managed Kyma runtime. CLI should be used with caution in such environment and user is prompted for confirmation.
func DetectManagedEnvironment(cmd *Command) error {
	detectStep := cmd.NewStep("Detecting managed Kyma runtime")

	if clusterinfo.IsManagedKyma(cmd.K8s.RestConfig()) {
		detectStep.LogWarn(managedKymaWarning)
		if !detectStep.PromptYesNo("Do you really want to proceed? ") {
			return errors.New("Command stopped by user")
		}
	}
	detectStep.Success()
	return nil
}
