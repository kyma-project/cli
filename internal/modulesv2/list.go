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

func (s *ListService) Run(ctx context.Context) ([]dtos.ListResult, []dtos.CommunityListResult, error) {
	if s.installedModulesRepository == nil {
		return nil, nil, nil
	}

	installedModules, err := s.installedModulesRepository.ListInstalledModules(ctx)
	if err != nil {
		return nil, nil, err
	}

	communityModules, err := s.installedModulesRepository.ListInstalledCommunityModules(ctx)
	if err != nil {
		return nil, nil, err
	}

	return dtos.ListResultsFromModuleInstallations(installedModules), dtos.CommunityListResultsFromInstallations(communityModules), nil
}

