package entities

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

var prohibitedNamespaces = []string{"kyma-system"}

type ExternalModuleTemplate struct {
	TemplateName   string
	ModuleName     string
	Version        string
	Namespace      string
	JsonDefinition string
}

func NewExternalModuleTemplateFromRaw(rawModuleTemplate *kyma.ModuleTemplate) *ExternalModuleTemplate {
	serializedTemplate, _ := json.Marshal(rawModuleTemplate)

	return &ExternalModuleTemplate{
		TemplateName:   rawModuleTemplate.Name,
		ModuleName:     rawModuleTemplate.Spec.ModuleName,
		Version:        rawModuleTemplate.Spec.Version,
		JsonDefinition: string(serializedTemplate),
	}
}

// SetNamespace saves information about a namespace in which the external module is going to be stored.
func (mt *ExternalModuleTemplate) SetNamespace(namespace string) error {
	if slices.Contains(prohibitedNamespaces, namespace) {
		return fmt.Errorf("'%s' namespace is not allowed", namespace)
	}

	mt.Namespace = namespace

	return nil
}
