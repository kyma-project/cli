package modulesv2

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type ListService struct {
	installedModulesRepository      repository.InstalledModulesRepository
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
		installationState, err := s.installationStateRepository.GetInstallationState(ctx, module.Status, module.Spec)
		if err != nil {
			return nil, err
		}

		results = append(results, dtos.ListResult{
			Name:                 module.Status.Name,
			Version:              module.Status.Version,
			Channel:              module.Status.Channel,
			State:                module.Status.State,
			Managed:              module.Spec.Managed == nil || *module.Spec.Managed,
			CustomResourcePolicy: module.Spec.CustomResourcePolicy,
			InstallationState:    installationState,
		})
	}

	return results, nil
}
