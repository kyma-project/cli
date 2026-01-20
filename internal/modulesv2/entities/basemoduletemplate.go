package entities

import (
	"encoding/json"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/out"
)

// BaseModuleTemplate is something like abstract class;
// Should be created only in "child" entities (coremoduletemplate, communitymoduletemplate)
type BaseModuleTemplate struct {
	ModuleName string
	Version    string

	name      string
	namespace string

	jsonSerializedTemplate string

	// not needed right now but will be needed in the future
	// data      *unstructured.Unstructured
	// manager   *kyma.Manager
}

func BaseModuleTemplateFromRaw(rawModuleTemplate *kyma.ModuleTemplate) *BaseModuleTemplate {
	entity := BaseModuleTemplate{}

	entity.ModuleName = rawModuleTemplate.Spec.ModuleName
	entity.Version = rawModuleTemplate.Spec.Version

	entity.name = rawModuleTemplate.GetName()
	entity.namespace = rawModuleTemplate.GetNamespace()

	entity.jsonSerializedTemplate = serialize(rawModuleTemplate)

	return &entity
}

func BaseModuleTemplateFromParams(templateName, moduleName, version, namespace string) *BaseModuleTemplate {
	return &BaseModuleTemplate{
		ModuleName: moduleName,
		Version:    version,
		name:       templateName,
		namespace:  namespace,
	}
}

func serialize(rawModuleTemplate *kyma.ModuleTemplate) string {
	serialized, err := json.Marshal(rawModuleTemplate)
	if err != nil {
		out.Errfln("faield to serialize moduletemplate: %v", rawModuleTemplate.GetName())
		out.Debugfln("moduletemplate: %v", rawModuleTemplate)

		return ""
	}

	return string(serialized)
}
