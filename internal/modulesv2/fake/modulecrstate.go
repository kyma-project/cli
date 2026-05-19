package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type ModuleCRStateRepository struct {
	GetModuleCRStateResult string
	GetModuleCRStateError  error
}

func (f *ModuleCRStateRepository) GetModuleCRState(_ context.Context, _ entities.ModuleInstallation) (string, error) {
	return f.GetModuleCRStateResult, f.GetModuleCRStateError
}
