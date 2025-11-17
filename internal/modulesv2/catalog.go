package modulesv2

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type Catalog struct {
	moduleTemplatesRepository *repository.ModuleTemplatesRepository
}

func NewCatalogService(moduleTemplatesRepository *repository.ModuleTemplatesRepository) *Catalog {
	return &Catalog{
		moduleTemplatesRepository: moduleTemplatesRepository,
	}
}

func (c *Catalog) Get(ctx context.Context) ([]entities.ModuleTemplate, error) {

	return nil, nil
}
