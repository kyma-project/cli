package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	"github.com/kyma-project/cli.v3/internal/out"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ModuleTemplatesRepository interface {
	ListCore(ctx context.Context) ([]*entities.CoreModuleTemplate, error)
	ListLocalCommunity(ctx context.Context) ([]*entities.CommunityModuleTemplate, error)
	ListExternalCommunity(ctx context.Context, urls []string, filterClause func(*entities.ExternalModuleTemplate) bool) ([]*entities.ExternalModuleTemplate, error)

	GetLocalCommunity(ctx context.Context, name, namespace string) (*entities.CommunityModuleTemplate, error)

	SaveCommunityModule(ctx context.Context, externalModule *entities.ExternalModuleTemplate) error
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

func (r *moduleTemplatesRepository) ListCore(ctx context.Context) ([]*entities.CoreModuleTemplate, error) {
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
		// support for legacy module templates
		legacyModuleTemplates, err := r.mapToCoreEntityLegacy(coreModuleTemplates)
		if err != nil {
			return nil, err
		}
		return legacyModuleTemplates, nil
	}

	return r.mapToCoreEntities(coreModuleTemplates, rawModulesReleaseMeta.Items), nil
}

func (r *moduleTemplatesRepository) ListLocalCommunity(ctx context.Context) ([]*entities.CommunityModuleTemplate, error) {
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

func (r *moduleTemplatesRepository) ListExternalCommunity(ctx context.Context, urls []string, filterClause func(*entities.ExternalModuleTemplate) bool) ([]*entities.ExternalModuleTemplate, error) {
	rawModuleTemplates, err := r.externalModuleTemplateRepository.Get(urls)
	if err != nil {
		return nil, err
	}

	communityEntities := r.mapToExternalCommunityEntities(rawModuleTemplates)

	if filterClause == nil {
		return communityEntities, nil
	}

	filteredCommunityEntities := []*entities.ExternalModuleTemplate{}
	for _, communityEntity := range communityEntities {
		if filterClause(communityEntity) {
			filteredCommunityEntities = append(filteredCommunityEntities, communityEntity)
		}
	}

	return filteredCommunityEntities, nil
}

func (r *moduleTemplatesRepository) GetLocalCommunity(ctx context.Context, name, namespace string) (*entities.CommunityModuleTemplate, error) {
	rawModuleTemplate, err := r.client.Kyma().GetModuleTemplate(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	return r.mapToCommunityEntity(rawModuleTemplate), nil
}

func (r *moduleTemplatesRepository) SaveCommunityModule(ctx context.Context, externalModule *entities.ExternalModuleTemplate) error {
	var kymaModuleTemplate kyma.ModuleTemplate
	err := json.Unmarshal([]byte(externalModule.JsonDefinition), &kymaModuleTemplate)
	if err != nil {
		return fmt.Errorf("failed to unmarshall %s moduleTemplate: %v", externalModule.ModuleName, err)
	}

	kymaModuleTemplate.Namespace = externalModule.Namespace

	unstructuredModule, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&kymaModuleTemplate)
	if err != nil {
		return err
	}

	return r.client.RootlessDynamic().Apply(ctx, &unstructured.Unstructured{Object: unstructuredModule}, false)
}

func (r *moduleTemplatesRepository) mapToCoreEntities(rawModuleTemplates []kyma.ModuleTemplate, rawReleaseMetas []kyma.ModuleReleaseMeta) []*entities.CoreModuleTemplate {
	entities := []*entities.CoreModuleTemplate{}

	for _, rawModuleTemplate := range rawModuleTemplates {
		assignments := getChannelVersionsAssignments(rawReleaseMetas, rawModuleTemplate.Spec.ModuleName, rawModuleTemplate.Spec.Version)
		for _, assignment := range assignments {
			entities = append(entities, r.mapToCoreEntity(&rawModuleTemplate, assignment.Channel))
		}
	}

	return entities
}

func (r *moduleTemplatesRepository) mapToCoreEntity(rawModuleTemplate *kyma.ModuleTemplate, channel string) *entities.CoreModuleTemplate {
	return entities.NewCoreModuleTemplateFromRaw(rawModuleTemplate, channel)
}

func (r *moduleTemplatesRepository) mapToCommunityEntities(rawModuleTemplates []kyma.ModuleTemplate) []*entities.CommunityModuleTemplate {
	entities := []*entities.CommunityModuleTemplate{}

	for _, rawModuleTemplate := range rawModuleTemplates {
		entities = append(entities, r.mapToCommunityEntity(&rawModuleTemplate))
	}

	return entities
}

func (r *moduleTemplatesRepository) mapToExternalCommunityEntities(rawModuleTemplates []kyma.ModuleTemplate) []*entities.ExternalModuleTemplate {
	extModuleTemplates := []*entities.ExternalModuleTemplate{}

	for _, rawModuleTemplate := range rawModuleTemplates {
		extModuleTemplates = append(extModuleTemplates, entities.NewExternalModuleTemplateFromRaw(&rawModuleTemplate))
	}

	return extModuleTemplates
}

func (r *moduleTemplatesRepository) mapToCommunityEntity(rawModuleTemplate *kyma.ModuleTemplate) *entities.CommunityModuleTemplate {
	return entities.NewCommunityModuleTemplateFromRaw(rawModuleTemplate)
}

func (r *moduleTemplatesRepository) mapToCoreEntityLegacy(coreModuleTemplates []kyma.ModuleTemplate) ([]*entities.CoreModuleTemplate, error) {
	coreModuleTemplateEntities := []*entities.CoreModuleTemplate{}

	for _, moduleTemplate := range coreModuleTemplates {
		type component struct {
			Version string `yaml:"version,omitempty"`
			Name    string `yaml:"name,omitempty"`
		}

		type descriptor struct {
			Component component `yaml:"component,omitempty"`
		}

		d := descriptor{}
		err := yaml.Unmarshal(moduleTemplate.Spec.Descriptor.Raw, &d)
		if err != nil {
			// unexpected error
			out.Debugfln("failed to parse %s module descriptor: %v", moduleTemplate.Spec.ModuleName, err)
			continue
		}

		nameElems := strings.Split(d.Component.Name, "/")
		componentName := nameElems[len(nameElems)-1]

		legacyModuleTemplate := entities.NewCoreModuleTemplateFromParams(moduleTemplate.Name, componentName, d.Component.Version, moduleTemplate.Spec.Channel, moduleTemplate.Namespace)
		coreModuleTemplateEntities = append(coreModuleTemplateEntities, legacyModuleTemplate)
	}

	if len(coreModuleTemplateEntities) != 0 {
		return coreModuleTemplateEntities, nil
	}

	return nil, errors.New("failed to list module catalog from the target Kyma environment")
}

func getChannelVersionsAssignments(rawReleaseMetas []kyma.ModuleReleaseMeta, moduleName, version string) []kyma.ChannelVersionAssignment {
	for _, rawReleaseMeta := range rawReleaseMetas {
		if rawReleaseMeta.Spec.ModuleName == moduleName {
			availableChannels := []kyma.ChannelVersionAssignment{}

			for _, channelAndVersion := range rawReleaseMeta.Spec.Channels {
				if channelAndVersion.Version == version {
					availableChannels = append(availableChannels, channelAndVersion)
				}
			}

			return availableChannels
		}
	}

	return []kyma.ChannelVersionAssignment{}
}

func isCommunityModule(moduleTemplate *kyma.ModuleTemplate) bool {
	managedBy, exist := moduleTemplate.ObjectMeta.Labels["operator.kyma-project.io/managed-by"]
	return !exist || managedBy != "kyma" || moduleTemplate.Namespace != "kyma-system"
}
