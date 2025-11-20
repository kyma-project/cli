package entities

import (
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type BaseModuleTemplate struct {
	ModuleName string
	Version    string

	name      string
	namespace string

	// not needed right now but will be needed in the future
	// data      *unstructured.Unstructured
	// manager   *kyma.Manager
}

func MapModuleTemplateFromRaw(rawModuleTemplate *kyma.ModuleTemplate) BaseModuleTemplate {
	entity := BaseModuleTemplate{}

	entity.ModuleName = rawModuleTemplate.Spec.ModuleName
	entity.Version = rawModuleTemplate.Spec.Version

	entity.name = rawModuleTemplate.GetName()
	entity.namespace = rawModuleTemplate.GetNamespace()

	return entity
}
