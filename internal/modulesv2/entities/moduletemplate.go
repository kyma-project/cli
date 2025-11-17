package entities

import (
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ModuleTemplate struct {
	name                string
	moduleName          string
	version             string
	channel             string // core-only
	namespace           string
	associatedResources []metav1.GroupVersionKind
	data                unstructured.Unstructured
	manager             *kyma.Manager
	source              string                      // community-only
	resourcesList       []unstructured.Unstructured // community-only
	isCommunity         bool
}

func MapModuleTemplateFromRaw(rawModuleTemplate *kyma.ModuleTemplate) ModuleTemplate {
	entity := ModuleTemplate{}

	entity.name = rawModuleTemplate.GetName()
	entity.moduleName = rawModuleTemplate.Spec.ModuleName
	entity.version = rawModuleTemplate.Spec.Version
	entity.namespace = rawModuleTemplate.GetNamespace()
	entity.associatedResources = rawModuleTemplate.Spec.AssociatedResources
	entity.data = rawModuleTemplate.Spec.Data
	entity.manager = rawModuleTemplate.Spec.Manager

	return entity
}

func (mt *ModuleTemplate) SetChannel(channel string) {
	mt.channel = channel
}
