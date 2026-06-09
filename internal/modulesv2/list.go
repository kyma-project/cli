package modulesv2

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
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

func (s *ListService) Run(ctx context.Context) ([]dtos.ListResult, []dtos.CommunityListResult, error) {
	var installedModules []entities.ModuleInstallation
	if s.installedModulesRepository != nil {
		var err error
		installedModules, err = s.installedModulesRepository.ListInstalledModules(ctx)
		if err != nil {
			return nil, nil, err
		}
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

	if s.installedModulesRepository == nil {
		return results, nil, nil
	}

	communityModules, err := s.installedModulesRepository.ListInstalledCommunityModules(ctx)
	if err != nil {
		return nil, nil, err
	}

	communityResults := make([]dtos.CommunityListResult, 0, len(communityModules))
	for _, module := range communityModules {
		communityResults = append(communityResults, dtos.CommunityListResult{
			Name:              module.FullName(),
			Version:           module.Version,
			ModuleState:       module.ModuleState,
			InstallationState: module.InstallationState,
		})
	}

	return results, communityResults, nil
}
