package kyma

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// copied from https://github.com/kyma-project/lifecycle-manager/tree/main/api/v1beta2
// we don't import the package here, because the lifecycle-manager/api package includes a lot of dependencies

// Module defines the components to be installed.
type Module struct {
	Name                 string `json:"name"`
	ControllerName       string `json:"controller,omitempty"`
	Channel              string `json:"channel,omitempty"`
	CustomResourcePolicy string `json:"customResourcePolicy,omitempty"`
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
