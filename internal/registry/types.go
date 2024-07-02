package registry

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	DockerRegistryGVR = schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "dockerregistries",
	}
)

type DockerRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            DockerRegistryStatus `json:"status,omitempty"`
}

type InternalAccess struct {
	SecretName string `json:"secretName,omitempty"`
}

type DockerRegistryStatus struct {
	State          string         `json:"state,omitempty"`
	Served         string         `json:"served"`
	InternalAccess InternalAccess `json:"internalAccess"`
}
