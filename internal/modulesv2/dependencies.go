package modulesv2

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/di"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
)

func SetupDIContainer(kymaConfig *cmdcommon.KymaConfig) (*di.DIContainer, error) {
	container := di.NewDIContainer()

	// 1. Register kube.Client - the foundation dependency
	di.RegisterTyped(container, func(c *di.DIContainer) (kube.Client, error) {
		return kymaConfig.GetKubeClient()
	})

	// 2. Register ExternalModuleTemplateRepository - has no dependencies
	di.RegisterTyped(container, func(c *di.DIContainer) (repository.ExternalModuleTemplateRepository, error) {
		return repository.NewExternalModuleTemplateRepository(), nil
	})

	// 3. Register ModuleTemplatesRepository - depends on kube.Client and ExternalModuleTemplateRepository
	di.RegisterTyped(container, func(c *di.DIContainer) (repository.ModuleTemplatesRepository, error) {
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

	// 4. Register Catalog - depends on ModuleTemplatesRepository
	di.RegisterTyped(container, func(c *di.DIContainer) (*CatalogService, error) {
		moduleRepo, err := di.GetTyped[repository.ModuleTemplatesRepository](c)
		if err != nil {
			return nil, err
		}

		return NewCatalogService(moduleRepo), nil
	})

	return container, nil
}
