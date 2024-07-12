package kyma

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// copied from https://github.com/kyma-project/lifecycle-manager/tree/main/api/v1beta2
// we don't import the package here, because the lifecycle-manager/api package includes a lot of dependencies

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
}

// KymaStatus defines the observed state of Kyma
type KymaStatus struct {
	Modules []ModuleStatus `json:"modules,omitempty"`
}

type ModuleStatus struct {
	Name    string `json:"name"`
	Channel string `json:"channel,omitempty"`
	Version string `json:"version,omitempty"`
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
)
