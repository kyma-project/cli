package repository

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type InstalledModulesRepository interface {
	ListInstalledModules(ctx context.Context) ([]kyma.ModuleStatus, error)
}

type installedModulesRepository struct {
	kymaClient kyma.Interface
}

func NewInstalledModulesRepository(kymaClient kyma.Interface) InstalledModulesRepository {
	return &installedModulesRepository{kymaClient: kymaClient}
}

func (r *installedModulesRepository) ListInstalledModules(ctx context.Context) ([]kyma.ModuleStatus, error) {
	kymaCR, err := r.kymaClient.GetDefaultKyma(ctx)
	if err != nil {
		return nil, err
	}

	return kymaCR.Status.Modules, nil
}
