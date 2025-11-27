package modulesv2

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type CatalogService struct {
	moduleTemplatesRepository repository.ModuleTemplatesRepository
	clusterMetadataRepository repository.ClusterMetadataRepository
}

func NewCatalogService(
	moduleTemplatesRepository repository.ModuleTemplatesRepository,
	clusterMetadataRepository repository.ClusterMetadataRepository,
) *CatalogService {
	return &CatalogService{
		moduleTemplatesRepository: moduleTemplatesRepository,
		clusterMetadataRepository: clusterMetadataRepository,
	}
}

func (c *CatalogService) Run(ctx context.Context, catalogConfig *dtos.CatalogConfig) ([]dtos.CatalogResult, error) {
	results := []dtos.CatalogResult{}

	if c.isClusterManagedByKLM(ctx) && catalogConfig.ListKyma {
		coreModules, err := c.moduleTemplatesRepository.ListCore(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list core modules: %v", err)
		}
		results = append(results, dtos.CatalogResultFromCoreModuleTemplates(coreModules)...)
	}

	if catalogConfig.ListCluster {
		localCommunityModules, err := c.moduleTemplatesRepository.ListLocalCommunity(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list local community modules: %v", err)
		}
		results = append(results, dtos.CatalogResultFromCommunityModuleTemplates(localCommunityModules)...)
	}

	externalCommunityModules, err := c.moduleTemplatesRepository.ListExternalCommunity(ctx, catalogConfig.ExternalUrls)
	if err != nil {
		return nil, fmt.Errorf("failed to list external community modules: %v", err)
	}
	results = append(results, dtos.CatalogResultFromCommunityModuleTemplates(externalCommunityModules)...)

	return results, nil
}

func (c *CatalogService) isClusterManagedByKLM(ctx context.Context) bool {
	clusterMetadata := c.clusterMetadataRepository.Get(ctx)
	return clusterMetadata.IsManagedByKLM
}
