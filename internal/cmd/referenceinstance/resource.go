package referenceinstance

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type KubernetesResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              KubernetesResourceSpec `json:"spec,omitempty"`
}

type KubernetesResourceSpec struct {
	Parameters   KubernetesResourceSpecParameters `json:"parameters,omitempty"`
	OfferingName string                           `json:"serviceOfferingName,omitempty"`
	PlanName     string                           `json:"servicePlanName,omitempty"`
}

type KubernetesResourceSpecParameters struct {
	InstanceID string        `json:"referenced_instance_id,omitempty"`
	Selectors  SpecSelectors `json:"selectors,omitempty"`
}

type SpecSelectors struct {
	InstanceLabelSelector []string `json:"instance_label_selector,omitempty"`
	InstanceNameSelector  string   `json:"instance_name_selector,omitempty"`
	PlanNameSelector      string   `json:"plan_name_selector,omitempty"`
}
