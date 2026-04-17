package modulesv2

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type ListService struct {
	installedModulesRepository repository.InstalledModulesRepository
}

func NewListService(installedModulesRepository repository.InstalledModulesRepository) *ListService {
	return &ListService{installedModulesRepository: installedModulesRepository}
}

func (s *ListService) Run(ctx context.Context) ([]dtos.ListResult, error) {
	installedModules, err := s.installedModulesRepository.ListInstalledModules(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]dtos.ListResult, 0, len(installedModules))
	for _, module := range installedModules {
		results = append(results, dtos.ListResult{
			Name:    module.Status.Name,
			Version: module.Status.Version,
			Channel: module.Status.Channel,
			State:   module.Status.State,
		})
	}

	return results, nil
}
