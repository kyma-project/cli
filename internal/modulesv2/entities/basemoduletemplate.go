package entities

import (
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

// BaseModuleTemplate is something like abstract class;
// Should be created only in "child" entities (coremoduletemplate, communitymoduletemplate)
type BaseModuleTemplate struct {
	ModuleName   string
	Version      string
	TemplateName string
	Namespace    string

	// not needed right now but will be needed in the future
	// data      *unstructured.Unstructured
	// manager   *kyma.Manager
}

func BaseModuleTemplateFromRaw(rawModuleTemplate *kyma.ModuleTemplate) *BaseModuleTemplate {
	entity := BaseModuleTemplate{}

	entity.ModuleName = rawModuleTemplate.Spec.ModuleName
	entity.Version = rawModuleTemplate.Spec.Version

	entity.TemplateName = rawModuleTemplate.GetName()
	entity.Namespace = rawModuleTemplate.GetNamespace()

	return &entity
}

func BaseModuleTemplateFromParams(templateName, moduleName, version, namespace string) *BaseModuleTemplate {
	return &BaseModuleTemplate{
		ModuleName:   moduleName,
		Version:      version,
		TemplateName: templateName,
		Namespace:    namespace,
	}
}
