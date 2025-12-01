package fake

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type ClusterMetadataRepository struct {
	IsManagedByKLM bool
}

func (m *ClusterMetadataRepository) Get(ctx context.Context) entities.ClusterMetadata {
	return entities.ClusterMetadata{IsManagedByKLM: m.IsManagedByKLM}
}
