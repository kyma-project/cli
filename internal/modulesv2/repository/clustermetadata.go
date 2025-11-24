package repository

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type ClusterMetadataRepository interface {
	Get(ctx context.Context) entities.ClusterMetadata
}

type clusterMetadataRepository struct {
	client kube.Client
}

func NewClusterMetadataRepository(client kube.Client) *clusterMetadataRepository {
	return &clusterMetadataRepository{client: client}
}

func (r *clusterMetadataRepository) Get(ctx context.Context) entities.ClusterMetadata {
	_, err := r.client.Kyma().GetDefaultKyma(ctx)
	if err != nil {
		return entities.ClusterMetadata{IsManagedByKLM: false}
	}

	return entities.ClusterMetadata{IsManagedByKLM: true}
}
