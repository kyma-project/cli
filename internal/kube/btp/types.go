package btp

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GVKServiceInstance = schema.GroupVersionKind{
		Group:   "services.cloud.sap.com",
		Version: "v1",
		Kind:    "ServiceInstance",
	}
	GVKServiceBinding = schema.GroupVersionKind{
		Group:   "services.cloud.sap.com",
		Version: "v1",
		Kind:    "ServiceBinding",
	}
)

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

const (
	ServicesAPIVersionV1 = "services.cloud.sap.com/v1"
	KindServiceInstance  = "ServiceInstance"
	KindServiceBinding   = "ServiceBinding"
)

type ServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceInstanceSpec `json:"spec,omitempty"`
}

type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceBindingSpec `json:"spec,omitempty"`
}

type ServiceInstanceSpec struct {
	Parameters           interface{} `json:"parameters,omitempty"`
	OfferingName         string      `json:"serviceOfferingName,omitempty"`
	PlanName             string      `json:"servicePlanName,omitempty"`
	BTPCredentialsSecret string      `json:"btpAccessCredentialsSecret,omitempty"`
}

type ServiceBindingSpec struct {
	Parameters          interface{} `json:"parameters,omitempty"`
	ServiceInstanceName string      `json:"serviceInstanceName,omitempty"`
	ExternalName        string      `json:"externalName,omitempty"`
	SecretName          string      `json:"secretName,omitempty"`
}
