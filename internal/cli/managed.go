package cli

import (
	"context"
	"errors"

	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/step"
)

const managedKymaWarning = "CAUTION: You are trying to use Kyma CLI to change a managed Kyma runtime (SAP BTP, Kyma runtime). This action may corrupt the existing installation. Proceed at your own risk."

// DetectManagedEnvironment introduces a step that checks if the target runtime is a managed Kyma runtime. CLI should be used with caution in such environment and user is prompted for confirmation.
func DetectManagedEnvironment(ctx context.Context, k kube.KymaKube, s step.Step) error {
	managed, err := clusterinfo.IsManagedKyma(ctx, k.RestConfig(), k.Static())
	if err != nil {
		return err
	}
	if managed {
		s.LogInfo(managedKymaWarning)
		if !s.PromptYesNo("Do you really want to proceed? ") {
			s.Failure()
			return errors.New("command stopped by user")
		}
	}
	s.Success()
	return nil
}
