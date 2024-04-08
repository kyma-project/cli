package referenceinstance

type KubernetesResourceSpecParameters struct {
	InstanceID string        `json:"referenced_instance_id,omitempty"`
	Selectors  SpecSelectors `json:"selectors,omitempty"`
}

type SpecSelectors struct {
	InstanceLabelSelector []string `json:"instance_label_selector,omitempty"`
	InstanceNameSelector  string   `json:"instance_name_selector,omitempty"`
	PlanNameSelector      string   `json:"plan_name_selector,omitempty"`
}
