package modulesv2

import (
	"errors"

	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/di"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

type ModuleOperations interface {
	Catalog() (*CatalogService, error)

	// TODO
	// Add() (*AddService, error)
	// Install() (*InstallService, error)
	// Pull() (*PullService, error)
	// etc.
}

type moduleOperations struct {
	kymaConfig *cmdcommon.KymaConfig
}

func NewModuleOperations(kymaConfig *cmdcommon.KymaConfig) *moduleOperations {
	return &moduleOperations{kymaConfig: kymaConfig}
}

func (m *moduleOperations) Catalog() (*CatalogService, error) {
	c := setupDIContainer(m.kymaConfig)

	catalogService, err := di.GetTyped[*CatalogService](c)
	if err != nil {
		return nil, errors.New("failed to execute the catalog command")
	}

	return catalogService, nil
}

func setupDIContainer(kymaConfig *cmdcommon.KymaConfig) *di.Container {
	container := di.NewContainer()

	di.RegisterTyped(container, func(c *di.Container) (kube.Client, error) {
		return kymaConfig.GetKubeClient()
	})

	di.RegisterTyped(container, func(c *di.Container) (repository.ExternalModuleTemplateRepository, error) {
		return repository.NewExternalModuleTemplateRepository(), nil
	})

	di.RegisterTyped(container, func(c *di.Container) (repository.ModuleTemplatesRepository, error) {
		kubeClient, err := di.GetTyped[kube.Client](c)
		if err != nil {
			return nil, err
		}

		externalRepo, err := di.GetTyped[repository.ExternalModuleTemplateRepository](c)
		if err != nil {
			return nil, err
		}

		return repository.NewModuleTemplatesRepository(kubeClient, externalRepo), nil
	})

	di.RegisterTyped(container, func(c *di.Container) (repository.ClusterMetadataRepository, error) {
		kubeClient, err := di.GetTyped[kube.Client](c)
		if err != nil {
			return nil, err
		}

		return repository.NewClusterMetadataRepository(kubeClient), nil
	})

	di.RegisterTyped(container, func(c *di.Container) (*CatalogService, error) {
		moduleRepo, err := di.GetTyped[repository.ModuleTemplatesRepository](c)
		if err != nil {
			return nil, err
		}

		metadataRepo, err := di.GetTyped[repository.ClusterMetadataRepository](c)
		if err != nil {
			return nil, err
		}

		return NewCatalogService(moduleRepo, metadataRepo), nil
	})

	return container
}
