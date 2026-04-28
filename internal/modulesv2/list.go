package modulesv2

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type ListService struct {
	installedModulesRepository  repository.InstalledModulesRepository
	installationStateRepository repository.ModuleInstallationStateRepository
}

func NewListService(installedModulesRepository repository.InstalledModulesRepository, installationStateRepository repository.ModuleInstallationStateRepository) *ListService {
	return &ListService{
		installedModulesRepository:  installedModulesRepository,
		installationStateRepository: installationStateRepository,
	}
}

func (s *ListService) Run(ctx context.Context) ([]dtos.ListResult, error) {
	installedModules, err := s.installedModulesRepository.ListInstalledModules(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]dtos.ListResult, 0, len(installedModules))
	for _, module := range installedModules {
		installationState, err := s.resolveInstallationState(ctx, module)
		if err != nil {
			return nil, err
		}

		results = append(results, dtos.ListResult{
			Name:                 module.Name,
			Version:              module.Version,
			Channel:              module.Channel,
			ModuleState:          module.ModuleState,
			Managed:              module.IsManaged(),
			CustomResourcePolicy: module.CustomResourcePolicy,
			InstallationState:    installationState,
		})
	}

	return results, nil
}

func (s *ListService) resolveInstallationState(ctx context.Context, module entities.ModuleInstallation) (string, error) {
	if module.CustomResourcePolicy == "CreateAndDelete" {
		return module.ModuleState, nil
	}

	if !module.IsManaged() {
		return module.ModuleState, nil
	}

	return s.installationStateRepository.GetInstallationState(ctx, module)
}
