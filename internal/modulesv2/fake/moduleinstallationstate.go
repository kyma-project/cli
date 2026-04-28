package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type ModuleInstallationStateRepository struct {
	GetInstallationStateResult string
	GetInstallationStateError  error
}

func (f *ModuleInstallationStateRepository) GetInstallationState(_ context.Context, _ entities.ModuleInstallation) (string, error) {
	return f.GetInstallationStateResult, f.GetInstallationStateError
}
