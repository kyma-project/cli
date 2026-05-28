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
	kymaClient        kyma.Interface
	moduleCRStateRepo ModuleCRStateRepository
}

func NewInstalledModulesRepository(kymaClient kyma.Interface, moduleCRStateRepo ModuleCRStateRepository) InstalledModulesRepository {
	return &installedModulesRepository{kymaClient: kymaClient, moduleCRStateRepo: moduleCRStateRepo}
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
		module, err := r.buildModule(ctx, raw)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}

	for _, spec := range kymaCR.Spec.Modules {
		if _, inStatus := statusByName[spec.Name]; inStatus {
			continue
		}
		module, err := r.buildModule(ctx, kyma.KymaModuleInfo{Spec: spec})
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}

	return modules, nil
}

func (r *installedModulesRepository) buildModule(ctx context.Context, raw kyma.KymaModuleInfo) (entities.ModuleInstallation, error) {
	module := entities.NewModuleInstallationFromRaw(raw)
	moduleState, err := r.moduleCRStateRepo.GetModuleCRState(ctx, *module)
	if err != nil {
		return entities.ModuleInstallation{}, err
	}
	module.ModuleState = moduleState
	return *module, nil
}
