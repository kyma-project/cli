package entities

import "fmt"

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

func (m *CommunityModuleTemplate) IsExternal() bool {
	return m.namespace == ""
}

func (m *CommunityModuleTemplate) GetNamespacedName() string {
	return fmt.Sprintf("%s/%s", m.namespace, m.name)
}
