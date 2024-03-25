package operator

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	GVRServiceInstance = schema.GroupVersionResource{
		Group:    "services.cloud.sap.com",
		Version:  "v1",
		Resource: "serviceinstances",
	}
	GVRServiceBinding = schema.GroupVersionResource{
		Group:    "services.cloud.sap.com",
		Version:  "v1",
		Resource: "servicebindings",
	}
)
