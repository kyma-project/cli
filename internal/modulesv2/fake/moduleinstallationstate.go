package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type ModuleInstallationStateRepository struct {
	GetInstallationStateResult string
	GetInstallationStateError  error
}

func (f *ModuleInstallationStateRepository) GetInstallationState(_ context.Context, _ kyma.ModuleStatus, _ kyma.Module) (string, error) {
	return f.GetInstallationStateResult, f.GetInstallationStateError
}
