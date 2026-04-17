package repository

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type InstalledModulesRepository interface {
	ListInstalledModules(ctx context.Context) ([]kyma.KymaModuleInfo, error)
}

type installedModulesRepository struct {
	kymaClient kyma.Interface
}

func NewInstalledModulesRepository(kymaClient kyma.Interface) InstalledModulesRepository {
	return &installedModulesRepository{kymaClient: kymaClient}
}

func (r *installedModulesRepository) ListInstalledModules(ctx context.Context) ([]kyma.KymaModuleInfo, error) {
	kymaCR, err := r.kymaClient.GetDefaultKyma(ctx)
	if err != nil {
		return nil, err
	}

	modules := make([]kyma.KymaModuleInfo, len(kymaCR.Status.Modules))
	for i, status := range kymaCR.Status.Modules {
		modules[i] = kyma.KymaModuleInfo{Status: status}
		for _, spec := range kymaCR.Spec.Modules {
			if spec.Name == status.Name {
				modules[i].Spec = spec
				break
			}
		}
	}
	return modules, nil
}
