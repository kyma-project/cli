package modulesv2

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type CatalogService struct {
	moduleTemplatesRepository repository.ModuleTemplatesRepository
}

func NewCatalogService(moduleTemplatesRepository repository.ModuleTemplatesRepository) *CatalogService {
	return &CatalogService{
		moduleTemplatesRepository: moduleTemplatesRepository,
	}
}

func (c *CatalogService) Run(ctx context.Context, urls []string) ([]dtos.CatalogResult, error) {
	results := []dtos.CatalogResult{}

	// todo: add support for clusters without kyma cr
	coreModules, err := c.moduleTemplatesRepository.ListCore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list core modules: %v", err)
	}

	localCommunityModules, err := c.moduleTemplatesRepository.ListLocalCommunity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list local community modules: %v", err)
	}

	externalCommunityModules, err := c.moduleTemplatesRepository.ListExternalCommunity(ctx, urls)
	if err != nil {
		return nil, fmt.Errorf("failed to list external community modules: %v", err)
	}

	results = append(results, dtos.CatalogResultFromCoreModuleTemplates(coreModules)...)
	results = append(results, dtos.CatalogResultFromCommunityModuleTemplates(localCommunityModules)...)
	results = append(results, dtos.CatalogResultFromCommunityModuleTemplates(externalCommunityModules)...)

	return results, nil
}
