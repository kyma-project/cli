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

func (m *CoreModuleTemplate) GetVersionWithChannel() string {
	return fmt.Sprintf("%s(%s)", m.Version, m.Channel)
}
