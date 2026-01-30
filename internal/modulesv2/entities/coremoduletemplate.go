package entities

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

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

func NewCoreModuleTemplateFromRaw(rawModuleTemplate *kyma.ModuleTemplate, channel string) *CoreModuleTemplate {
	moduleTemplateEntity := BaseModuleTemplateFromRaw(rawModuleTemplate)

	return NewCoreModuleTemplate(moduleTemplateEntity, channel)
}

func NewCoreModuleTemplateFromParams(templateName, moduleName, version, channel, namespace string) *CoreModuleTemplate {
	base := BaseModuleTemplateFromParams(templateName, moduleName, version, namespace)

	return &CoreModuleTemplate{
		*base,
		channel,
	}
}

func (m *CoreModuleTemplate) GetVersionWithChannel() string {
	return fmt.Sprintf("%s(%s)", m.Version, m.Channel)
}
