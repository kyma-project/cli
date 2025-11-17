package repository

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
)

type ModuleTemplatesRepository interface {
	ListCore(ctx context.Context) ([]entities.ModuleTemplate, error)
	ListCommunity(ctx context.Context) ([]entities.ModuleTemplate, error)
}

type moduleTemplatesRepository struct {
	client kube.Client
}

func NewModuleTemplatesRepository(client kube.Client) *moduleTemplatesRepository {
	return &moduleTemplatesRepository{
		client: client,
	}
}

func (r *moduleTemplatesRepository) getLocal(ctx context.Context) ([]kyma.ModuleTemplate, error) {
	moduleTemplates, err := r.client.Kyma().ListModuleTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list module templates: %v", err)
	}

	return moduleTemplates.Items, nil
}

func (r *moduleTemplatesRepository) ListCore(ctx context.Context) ([]entities.ModuleTemplate, error) {
	rawModuleTemplates, err := r.getLocal(ctx)
	if err != nil {
		return nil, err
	}

	rawModulesReleaseMeta, err := r.client.Kyma().ListModuleReleaseMeta(ctx)
	if err != nil {
		// TODO: add support for legacy catalog
		return nil, err
	}

	return r.mapToCoreEntities(rawModuleTemplates, rawModulesReleaseMeta.Items), nil
}

func (r *moduleTemplatesRepository) ListCommunity(ctx context.Context) ([]entities.ModuleTemplate, error) {

}

func (r *moduleTemplatesRepository) mapToCoreEntities(rawModuleTemplates []kyma.ModuleTemplate, rawReleaseMetas []kyma.ModuleReleaseMeta) []entities.ModuleTemplate {
	entities := make([]entities.ModuleTemplate, 0)

	for _, rawModuleTemplate := range rawModuleTemplates {
		assignments := getChannelVersionsAssignments(rawReleaseMetas, rawModuleTemplate.Name)
		for _, assignment := range assignments {
			entities = append(entities, r.mapToCoreEntity(&rawModuleTemplate, assignment.Channel))
		}
	}

	return entities
}

func (r *moduleTemplatesRepository) mapToCoreEntity(rawModuleTemplate *kyma.ModuleTemplate, channel string) entities.ModuleTemplate {
	moduleTemplateEntity := entities.MapModuleTemplateFromRaw(rawModuleTemplate)
	moduleTemplateEntity.SetChannel(channel)

	return moduleTemplateEntity
}

func getChannelVersionsAssignments(rawReleaseMetas []kyma.ModuleReleaseMeta, moduleName string) []kyma.ChannelVersionAssignment {
	for _, rawReleaseMeta := range rawReleaseMetas {
		if rawReleaseMeta.Spec.ModuleName == moduleName {
			return rawReleaseMeta.Spec.Channels
		}
	}

	return []kyma.ChannelVersionAssignment{}
}
