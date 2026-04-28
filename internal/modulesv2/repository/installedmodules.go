package repository

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type InstalledModulesRepository interface {
	ListInstalledModules(ctx context.Context) ([]entities.ModuleInstallation, error)
}

type installedModulesRepository struct {
	kymaClient kyma.Interface
}

func NewInstalledModulesRepository(kymaClient kyma.Interface) InstalledModulesRepository {
	return &installedModulesRepository{kymaClient: kymaClient}
}

func (r *installedModulesRepository) ListInstalledModules(ctx context.Context) ([]entities.ModuleInstallation, error) {
	kymaCR, err := r.kymaClient.GetDefaultKyma(ctx)
	if err != nil {
		return nil, err
	}

	modules := make([]entities.ModuleInstallation, len(kymaCR.Status.Modules))
	for i, status := range kymaCR.Status.Modules {
		raw := kyma.KymaModuleInfo{Status: status}
		for _, spec := range kymaCR.Spec.Modules {
			if spec.Name == status.Name {
				raw.Spec = spec
				break
			}
		}
		modules[i] = *entities.NewModuleInstallationFromRaw(raw)
	}
	return modules, nil
}
