package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// ToUnstructured converts the given data to an Unstructured object
func ToUnstructured(requestData interface{}, gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&requestData)
	if err != nil {
		return nil, err
	}

	unstructuredObj := &unstructured.Unstructured{Object: u}
	unstructuredObj.SetGroupVersionKind(gvk)

	return unstructuredObj, nil
}
