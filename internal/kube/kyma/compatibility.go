package kyma

import (
	"fmt"
)

// Compatibility, remove when old moduleTemplate definitions are removed
// getOldModuleTemplate returns matching ModuleTemplate from list, based on old ModuleTemplate definitions
func getOldModuleTemplate(moduleTemplates *ModuleTemplateList, moduleName, moduleChannel string) *ModuleTemplate {
	for _, moduleTemplate := range moduleTemplates.Items {
		// old module templates have name in moduleName-moduleChannel format
		if moduleTemplate.ObjectMeta.Name == fmt.Sprintf("%s-%s", moduleName, moduleChannel) {
			return &moduleTemplate
		}
	}
	return nil
}
