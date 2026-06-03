package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestExtractStateFromObject_ReturnsStateField(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"state": "Ready",
			},
		},
	}

	state := extractStateFromObject(obj)

	require.Equal(t, "Ready", state)
}

func TestExtractStateFromObject_ReturnsStateFromConditions(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Available",
						"status": "True",
					},
				},
			},
		},
	}

	state := extractStateFromObject(obj)

	require.Equal(t, "Ready", state)
}

func TestExtractStateFromObject_ReturnsStateFromReplicas(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"replicas": int64(3),
			},
			"status": map[string]interface{}{
				"readyReplicas": int64(3),
			},
		},
	}

	state := extractStateFromObject(obj)

	require.Equal(t, "Ready", state)
}

func TestExtractStateFromObject_ReturnsEmptyWhenNoStatus(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}

	state := extractStateFromObject(obj)

	require.Equal(t, "", state)
}
