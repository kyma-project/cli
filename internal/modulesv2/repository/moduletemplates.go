package repository

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type ModuleTemplatesRepository interface {
	ListCore(ctx context.Context) ([]entities.CoreModuleTemplate, error)
	ListLocalCommunity(ctx context.Context) ([]entities.CommunityModuleTemplate, error)
	ListExternalCommunity(ctx context.Context, urls []string) ([]entities.CommunityModuleTemplate, error)
}

type moduleTemplatesRepository struct {
	client                           kube.Client
	externalModuleTemplateRepository ExternalModuleTemplateRepository
}

func NewModuleTemplatesRepository(client kube.Client, externalModuleTemplateRepository ExternalModuleTemplateRepository) *moduleTemplatesRepository {
	return &moduleTemplatesRepository{
		client:                           client,
		externalModuleTemplateRepository: externalModuleTemplateRepository,
	}
}

func (r *moduleTemplatesRepository) getLocal(ctx context.Context) ([]kyma.ModuleTemplate, error) {
	moduleTemplates, err := r.client.Kyma().ListModuleTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list module templates: %v", err)
	}

	return moduleTemplates.Items, nil
}

func (r *moduleTemplatesRepository) ListCore(ctx context.Context) ([]entities.CoreModuleTemplate, error) {
	rawModuleTemplates, err := r.getLocal(ctx)
	if err != nil {
		return nil, err
	}

	coreModuleTemplates := []kyma.ModuleTemplate{}
	for _, moduleTemplate := range rawModuleTemplates {
		if !isCommunityModule(&moduleTemplate) {
			coreModuleTemplates = append(coreModuleTemplates, moduleTemplate)
		}
	}

	rawModulesReleaseMeta, err := r.client.Kyma().ListModuleReleaseMeta(ctx)
	if err != nil {
		// TODO: add support for legacy catalog
		return nil, err
	}

	return r.mapToCoreEntities(coreModuleTemplates, rawModulesReleaseMeta.Items), nil
}

func (r *moduleTemplatesRepository) ListLocalCommunity(ctx context.Context) ([]entities.CommunityModuleTemplate, error) {
	rawModuleTemplates, err := r.getLocal(ctx)
	if err != nil {
		return nil, err
	}

	communityModuleTemplates := []kyma.ModuleTemplate{}
	for _, moduleTemplate := range rawModuleTemplates {
		if isCommunityModule(&moduleTemplate) {
			communityModuleTemplates = append(communityModuleTemplates, moduleTemplate)
		}
	}

	return r.mapToCommunityEntities(communityModuleTemplates), nil
}

func (r *moduleTemplatesRepository) ListExternalCommunity(ctx context.Context, urls []string) ([]entities.CommunityModuleTemplate, error) {
	rawModuleTemplates, err := r.externalModuleTemplateRepository.Get(urls)
	if err != nil {
		return nil, err
	}

	return r.mapToCommunityEntities(rawModuleTemplates), nil
}

func (r *moduleTemplatesRepository) mapToCoreEntities(rawModuleTemplates []kyma.ModuleTemplate, rawReleaseMetas []kyma.ModuleReleaseMeta) []entities.CoreModuleTemplate {
	entities := []entities.CoreModuleTemplate{}

	for _, rawModuleTemplate := range rawModuleTemplates {
		assignments := getChannelVersionsAssignments(rawReleaseMetas, rawModuleTemplate.Spec.ModuleName)
		for _, assignment := range assignments {
			entities = append(entities, r.mapToCoreEntity(&rawModuleTemplate, assignment.Channel))
		}
	}

	return entities
}

func (r *moduleTemplatesRepository) mapToCoreEntity(rawModuleTemplate *kyma.ModuleTemplate, channel string) entities.CoreModuleTemplate {
	moduleTemplateEntity := entities.MapModuleTemplateFromRaw(rawModuleTemplate)

	return *entities.NewCoreModuleTemplate(&moduleTemplateEntity, channel)
}

func (r *moduleTemplatesRepository) mapToCommunityEntities(rawModuleTemplates []kyma.ModuleTemplate) []entities.CommunityModuleTemplate {
	entities := []entities.CommunityModuleTemplate{}

	for _, rawModuleTemplate := range rawModuleTemplates {
		entities = append(entities, r.mapToCommunityEntity(&rawModuleTemplate))
	}

	return entities
}

func (r *moduleTemplatesRepository) mapToCommunityEntity(rawModuleTemplate *kyma.ModuleTemplate) entities.CommunityModuleTemplate {
	moduleTemplateEntity := entities.MapModuleTemplateFromRaw(rawModuleTemplate)
	sourceURL := rawModuleTemplate.Annotations["source"]
	resources := map[string]string{}

	for _, rawResource := range rawModuleTemplate.Spec.Resources {
		key := rawResource.Name
		value := rawResource.Link

		resources[key] = value
	}

	return *entities.NewCommunityModuleTemplate(&moduleTemplateEntity, sourceURL, resources)
}

func getChannelVersionsAssignments(rawReleaseMetas []kyma.ModuleReleaseMeta, moduleName string) []kyma.ChannelVersionAssignment {
	for _, rawReleaseMeta := range rawReleaseMetas {
		if rawReleaseMeta.Spec.ModuleName == moduleName {
			return rawReleaseMeta.Spec.Channels
		}
	}

	return []kyma.ChannelVersionAssignment{}
}

func isCommunityModule(moduleTemplate *kyma.ModuleTemplate) bool {
	managedBy, exist := moduleTemplate.ObjectMeta.Labels["operator.kyma-project.io/managed-by"]
	return !exist || managedBy != "kyma" || moduleTemplate.Namespace != "kyma-system"
}
