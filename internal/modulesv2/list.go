package modulesv2

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type ListService struct {
	installedModulesRepository repository.ModuleInstallationsRepository
}

func NewListService(installedModulesRepository repository.ModuleInstallationsRepository) *ListService {
	return &ListService{
		installedModulesRepository: installedModulesRepository,
	}
}

func (s *ListService) Run(ctx context.Context) ([]dtos.ListResult, error) {
	installedModules, err := s.installedModulesRepository.ListInstalledModules(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]dtos.ListResult, 0, len(installedModules))
	for _, module := range installedModules {
		results = append(results, dtos.ListResult{
			Name:                 module.Name,
			Version:              module.Version,
			Channel:              module.Channel,
			ModuleState:          module.ModuleState,
			Managed:              module.IsManaged(),
			CustomResourcePolicy: module.CustomResourcePolicy,
			InstallationState:    module.InstallationState,
		})
	}

	return results, nil
}
