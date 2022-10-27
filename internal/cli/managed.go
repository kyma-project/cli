package cli

import (
	"errors"

	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
)

const managedKymaWarning = "CAUTION: You are trying to use Kyma CLI to change a managed Kyma runtime (SAP Kyma Runtime). This action may corrupt the Kyma runtime. Proceed at your own risk."

// DetectManagedEnvironment introduces a step that checks if the target runtime is a managed Kyma runtime. CLI should be used with caution in such environment and user is prompted for confirmation.
func DetectManagedEnvironment(k kube.KymaKube, s step.Step) error {
	if clusterinfo.IsManagedKyma(k.RestConfig()) {
		s.LogWarn(managedKymaWarning)
		if !s.PromptYesNo("Do you really want to proceed? ") {
			return errors.New("Command stopped by user")
		}
	}
	return nil
}
