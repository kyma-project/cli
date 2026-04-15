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
	return []dtos.ListResult{}, nil
}
