package operator

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	GVKServiceInstance = schema.GroupVersionKind{
		Group:   "services.cloud.sap.com",
		Version: "v1",
		Kind:    "ServiceInstance",
	}
)
