package kyma

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetDefaultKyma(ctx context.Context, client kube.Client) (*unstructured.Unstructured, error) {
	return client.Dynamic().Resource(GVRKyma).
		Namespace(("kyma-system")).
		Get(ctx, "default", metav1.GetOptions{})
}

func UpdateDefaultKyma(ctx context.Context, client kube.Client, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return client.Dynamic().Resource(GVRKyma).
		Namespace(("kyma-system")).
		Update(ctx, obj, metav1.UpdateOptions{})
}

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
