package repository

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func extractStateFromObject(obj *unstructured.Unstructured) string {
	statusRaw, ok := obj.Object["status"]
	if !ok || statusRaw == nil {
		return ""
	}
	status := statusRaw.(map[string]any)
	if state, ok := status["state"]; ok {
		return state.(string)
	}

	if conditions, ok := status["conditions"]; ok {
		return getStateFromConditions(conditions.([]any))
	}

	if readyReplicas, ok := status["readyReplicas"]; ok {
		spec := obj.Object["spec"].(map[string]any)
		if wantedReplicas, ok := spec["replicas"]; ok {
			return resolveStateFromReplicas(readyReplicas.(int64), wantedReplicas.(int64))
		}
	}

	return ""
}

func getStateFromConditions(conditions []interface{}) string {
	for _, condition := range conditions {
		c := condition.(map[string]interface{})
		if c["status"] != "True" {
			continue
		}
		switch c["type"].(string) {
		case "Available":
			return "Ready"
		case "Processing", "Error", "Warning":
			return c["type"].(string)
		}
	}
	return ""
}

func resolveStateFromReplicas(ready, wanted int64) string {
	if ready == wanted {
		return "Ready"
	}
	if ready < wanted {
		return "Processing"
	}
	return "Deleting"
}
