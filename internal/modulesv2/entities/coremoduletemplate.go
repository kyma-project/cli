package entities

import "fmt"

type CoreModuleTemplate struct {
	BaseModuleTemplate
	Channel string
}

func NewCoreModuleTemplate(base *BaseModuleTemplate, channel string) *CoreModuleTemplate {
	return &CoreModuleTemplate{
		*base,
		channel,
	}
}

func NewCoreModuleTemplateFromParams(templateName, moduleName, version, channel, namespace string) *CoreModuleTemplate {
	base := MapBaseModuleTemplateFromParams(templateName, moduleName, version, namespace)

	return &CoreModuleTemplate{
		*base,
		channel,
	}
}

func (m *CoreModuleTemplate) GetVersionWithChannel() string {
	return fmt.Sprintf("%s(%s)", m.Version, m.Channel)
}
