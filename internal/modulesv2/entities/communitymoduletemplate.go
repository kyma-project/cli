package entities

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type CommunityModuleTemplate struct {
	BaseModuleTemplate
	sourceURL string
	resources map[string]string
}

func NewCommunityModuleTemplate(base *BaseModuleTemplate, sourceURL string, resources map[string]string) *CommunityModuleTemplate {
	return &CommunityModuleTemplate{
		*base,
		sourceURL,
		resources,
	}
}

func NewCommunityModuleTemplateFromRaw(rawModuleTemplate *kyma.ModuleTemplate) *CommunityModuleTemplate {
	moduleTemplateEntity := BaseModuleTemplateFromRaw(rawModuleTemplate)
	sourceURL := rawModuleTemplate.Annotations["source"]
	resources := map[string]string{}

	for _, rawResource := range rawModuleTemplate.Spec.Resources {
		key := rawResource.Name
		value := rawResource.Link

		resources[key] = value
	}

	return NewCommunityModuleTemplate(moduleTemplateEntity, sourceURL, resources)
}

func (m *CommunityModuleTemplate) IsExternal() bool {
	return m.namespace == ""
}

func (m *CommunityModuleTemplate) GetNamespacedName() string {
	return fmt.Sprintf("%s/%s", m.namespace, m.name)
}
