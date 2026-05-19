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

	statusByName := make(map[string]kyma.ModuleStatus, len(kymaCR.Status.Modules))
	for _, status := range kymaCR.Status.Modules {
		statusByName[status.Name] = status
	}

	specByName := make(map[string]kyma.Module, len(kymaCR.Spec.Modules))
	for _, spec := range kymaCR.Spec.Modules {
		specByName[spec.Name] = spec
	}

	var modules []entities.ModuleInstallation

	for _, status := range kymaCR.Status.Modules {
		raw := kyma.KymaModuleInfo{Status: status}
		if spec, ok := specByName[status.Name]; ok {
			raw.Spec = spec
		}
		modules = append(modules, *entities.NewModuleInstallationFromRaw(raw))
	}

	for _, spec := range kymaCR.Spec.Modules {
		if _, inStatus := statusByName[spec.Name]; inStatus {
			continue
		}
		raw := kyma.KymaModuleInfo{Spec: spec}
		modules = append(modules, *entities.NewModuleInstallationFromRaw(raw))
	}

	return modules, nil
}
