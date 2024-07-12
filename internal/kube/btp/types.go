package btp

import (
	"k8s.io/apimachinery/pkg/api/meta"
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
	Status            CommonStatus        `json:"status,omitempty"`
}

type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceBindingSpec `json:"spec,omitempty"`
	Status            CommonStatus       `json:"status"`
}

type CommonStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Ready      string             `json:"ready,omitempty"`
	InstanceID string             `json:"instanceID,omitempty"`
}

type ServiceInstanceSpec struct {
	Parameters                 interface{} `json:"parameters,omitempty"`
	ServiceOfferingName        string      `json:"serviceOfferingName,omitempty"`
	ServicePlanName            string      `json:"servicePlanName,omitempty"`
	BTPAccessCredentialsSecret string      `json:"btpAccessCredentialsSecret,omitempty"`
	ExternalName               string      `json:"externalName,omitempty"`
}

type ServiceBindingSpec struct {
	Parameters          interface{} `json:"parameters,omitempty"`
	ServiceInstanceName string      `json:"serviceInstanceName,omitempty"`
	ExternalName        string      `json:"externalName,omitempty"`
	SecretName          string      `json:"secretName,omitempty"`
}

// IsReady returns readiness status
func (s *CommonStatus) IsReady() bool {
	return (s.Ready == "True") &&
		isConditionTrue(s.Conditions, "Succeeded") &&
		isConditionTrue(s.Conditions, "Ready")
}

// IsFailed returns if at least one condition has Failed status
func (s *CommonStatus) IsFailed() bool {
	return (s.Ready == "False") &&
		isConditionTrue(s.Conditions, "Failed")
}

// GetConditionMessage returns message of condition in given type
func (s *CommonStatus) GetConditionMessage(conditionType string) string {
	condition := meta.FindStatusCondition(s.Conditions, conditionType)
	if condition == nil {
		return ""
	}
	return condition.Message
}

func isConditionTrue(conditions []metav1.Condition, conditionType string) bool {
	condition := meta.FindStatusCondition(conditions, conditionType)
	return condition != nil && condition.Status == metav1.ConditionTrue
}
