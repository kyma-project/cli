package unstructured

import (
	"fmt"
	"github.com/kyma-project/cli/internal/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	triggerApiVersion = "eventing.knative.dev/v1alpha1"
	triggerNameFormat = "%s-%s"
)

func NewTriggers(cfg workspace.Cfg) []unstructured.Unstructured {
	var list []unstructured.Unstructured
	for _, trigger := range cfg.Triggers {
		out := unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": triggerApiVersion,
			"kind":       "Trigger",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf(triggerNameFormat, cfg.Name, trigger.Source),
				"namespace": cfg.Namespace,
				"labels":    cfg.Labels,
			},
			"spec": map[string]interface{}{
				"broker": "default",
				"filter": map[string]interface{}{
					"attributes": map[string]interface{}{
						"eventtypeversion": trigger.EventTypeVersion,
						"source":           trigger.Source,
						"type":             trigger.Type,
					},
				},
				"subscriber": map[string]interface{}{
					"ref": map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Service",
						"name":       cfg.Name,
						"namespace":  cfg.Namespace,
					},
				},
			},
		}}
		list = append(list, out)
	}

	return list
}
