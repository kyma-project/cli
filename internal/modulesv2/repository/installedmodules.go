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
	kymaClient            kyma.Interface
	moduleCRStateRepo     ModuleCRStateRepository
	installationStateRepo ModuleInstallationStateRepository
}

func NewInstalledModulesRepository(kymaClient kyma.Interface, moduleCRStateRepo ModuleCRStateRepository, installationStateRepo ModuleInstallationStateRepository) InstalledModulesRepository {
	return &installedModulesRepository{kymaClient: kymaClient, moduleCRStateRepo: moduleCRStateRepo, installationStateRepo: installationStateRepo}
}

func (r *installedModulesRepository) ListInstalledModules(ctx context.Context) ([]entities.ModuleInstallation, error) {
	kymaCR, err := r.kymaClient.GetDefaultKyma(ctx)
	if err != nil {
		return nil, err
	}
	return r.resolveInstalledModules(ctx, kymaCR.Spec.Modules, kymaCR.Status.Modules)
}

func (r *installedModulesRepository) resolveInstalledModules(ctx context.Context, specs []kyma.Module, statuses []kyma.ModuleStatus) ([]entities.ModuleInstallation, error) {
	modules, err := r.buildModulesFromStatuses(ctx, statuses, specs)
	if err != nil {
		return nil, err
	}

	specOnly, err := r.buildModulesFromSpecsOnly(ctx, specs, statuses)
	if err != nil {
		return nil, err
	}

	return append(modules, specOnly...), nil
}

func (r *installedModulesRepository) buildModulesFromStatuses(ctx context.Context, statuses []kyma.ModuleStatus, specs []kyma.Module) ([]entities.ModuleInstallation, error) {
	specByName := make(map[string]kyma.Module, len(specs))
	for _, spec := range specs {
		specByName[spec.Name] = spec
	}

	var modules []entities.ModuleInstallation
	for _, status := range statuses {
		module, err := r.buildModule(ctx, kyma.KymaModuleInfo{Status: status, Spec: specByName[status.Name]})
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}
	return modules, nil
}

func (r *installedModulesRepository) buildModulesFromSpecsOnly(ctx context.Context, specs []kyma.Module, statuses []kyma.ModuleStatus) ([]entities.ModuleInstallation, error) {
	statusByName := make(map[string]kyma.ModuleStatus, len(statuses))
	for _, status := range statuses {
		statusByName[status.Name] = status
	}

	var modules []entities.ModuleInstallation
	for _, spec := range specs {
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
	installationState, err := r.resolveInstallationState(ctx, *module)
	if err != nil {
		return entities.ModuleInstallation{}, err
	}
	module.InstallationState = installationState
	return *module, nil
}

func (r *installedModulesRepository) resolveInstallationState(ctx context.Context, module entities.ModuleInstallation) (string, error) {
	if module.IsBeingDeleted() {
		return module.KymaModuleState, nil
	}

	if module.CustomResourcePolicy == "CreateAndDelete" {
		return module.KymaModuleState, nil
	}

	if !module.IsManaged() {
		return module.KymaModuleState, nil
	}

	return r.installationStateRepo.GetInstallationState(ctx, module)
}
