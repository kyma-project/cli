package kyma

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// copied from https://github.com/kyma-project/lifecycle-manager/tree/main/api/v1beta2
// we don't import the package here, because the lifecycle-manager/api package includes a lot of dependencies

// ModuleReleaseMetaList contains a list of ModuleReleaseMeta.
type ModuleReleaseMetaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModuleReleaseMeta `json:"items"`
}

type ModuleReleaseMeta struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ModuleReleaseMetaSpec `json:"spec,omitempty"`
}

// ModuleReleaseMetaSpec defines the channel-version assignments for a module.
type ModuleReleaseMetaSpec struct {
	ModuleName string                     `json:"moduleName"`
	Channels   []ChannelVersionAssignment `json:"channels"`
}

type ChannelVersionAssignment struct {
	Channel string `json:"channel"`
	Version string `json:"version"`
}

// ModuleTemplateList contains a list of ModuleTemplate.
type ModuleTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModuleTemplate `json:"items"`
}

type ModuleTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ModuleTemplateSpec `json:"spec,omitempty"`
}

// ModuleTemplateSpec defines the desired state of ModuleTemplate.
type ModuleTemplateSpec struct {
	Channel             string                    `json:"channel"`
	Version             string                    `json:"version"`
	ModuleName          string                    `json:"moduleName"`
	Mandatory           bool                      `json:"mandatory"`
	Data                unstructured.Unstructured `json:"data,omitempty"`
	Descriptor          runtime.RawExtension      `json:"descriptor"`
	CustomStateCheck    []CustomStateCheck        `json:"customStateCheck,omitempty"`
	Resources           []Resource                `json:"resources,omitempty"`
	Info                ModuleInfo                `json:"info,omitempty"`
	AssociatedResources []metav1.GroupVersionKind `json:"associatedResources,omitempty"`
	Manager             Manager                   `json:"manager,omitempty"`
}

// Manager defines the structure for the manager field in ModuleTemplateSpec.
type Manager struct {
	metav1.GroupVersionKind `json:",inline"`
	Namespace               string `json:"namespace,omitempty"`
	Name                    string `json:"name"`
}

type ModuleInfo struct {
	Repository    string       `json:"repository"`
	Documentation string       `json:"documentation"`
	Icons         []ModuleIcon `json:"icons,omitempty"`
}

type ModuleIcon struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type CustomStateCheck struct {
	JSONPath    string `json:"jsonPath" yaml:"jsonPath"`
	Value       string `json:"value" yaml:"value"`
	MappedState string `json:"mappedState" yaml:"mappedState"`
}

type Resource struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// Kyma is the Schema for the kymas API.
type Kyma struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KymaSpec   `json:"spec,omitempty"`
	Status KymaStatus `json:"status,omitempty"`
}

// KymaSpec defines the desired state of Kyma.
type KymaSpec struct {
	Channel string   `json:"channel"`
	Modules []Module `json:"modules,omitempty"`
}

// Module defines the components to be installed.
type Module struct {
	Name                 string `json:"name"`
	ControllerName       string `json:"controller,omitempty"`
	Channel              string `json:"channel,omitempty"`
	CustomResourcePolicy string `json:"customResourcePolicy,omitempty"`
	Managed              bool   `json:"managed,omitempty"`
}

// KymaStatus defines the observed state of Kyma
type KymaStatus struct {
	Modules []ModuleStatus `json:"modules,omitempty"`
}

type ModuleStatus struct {
	Name    string `json:"name"`
	Channel string `json:"channel,omitempty"`
	Version string `json:"version,omitempty"`
	State   string `json:"state,omitempty"`
}

// ModuleFromInterface converts a map retrieved from the Unstructured kyma CR to a Module struct.
func ModuleFromInterface(i map[string]interface{}) Module {
	module := Module{Name: i["name"].(string)}
	if i["controllerName"] != nil {
		module.ControllerName = i["controllerName"].(string)
	}
	if i["channel"] != nil {
		module.Channel = i["channel"].(string)
	}
	if i["customResourcePolicy"] != nil {
		module.CustomResourcePolicy = i["customResourcePolicy"].(string)
	}
	return module
}

var (
	GVRKyma = schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta2",
		Resource: "kymas",
	}

	GVRModuleTemplate = schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta2",
		Resource: "moduletemplates",
	}

	GVRModuleReleaseMeta = schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta2",
		Resource: "modulereleasemetas",
	}
)
