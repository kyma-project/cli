package modules

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
)

var (
	testModuleTemplate1 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "serverless-1",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "serverless",
				"version":    "0.0.1",
				"info": map[string]interface{}{
					"repository": "url-1",
				},
			},
		},
	}

	testModuleTemplate2 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "serverless-2",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "serverless",
				"version":    "0.0.2",
				"info": map[string]interface{}{
					"repository": "url-2",
				},
			},
		},
	}

	testModuleTemplate3 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "keda-1",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "keda",
				"version":    "0.1",
				"info": map[string]interface{}{
					"repository": "url-3",
				},
			},
		},
	}

	testModuleTemplate4 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleTemplate",
			"metadata": map[string]interface{}{
				"name":      "keda-2",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "keda",
				"version":    "0.2",
			},
		},
	}

	testReleaseMeta1 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleReleaseMeta",
			"metadata": map[string]interface{}{
				"name":      "serverless",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "serverless",
				"channels": []interface{}{
					map[string]interface{}{
						"version": "0.0.1",
						"channel": "fast",
					},
				},
			},
		},
	}

	testReleaseMeta2 = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "ModuleReleaseMeta",
			"metadata": map[string]interface{}{
				"name":      "keda",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"moduleName": "keda",
				"channels": []interface{}{
					map[string]interface{}{
						"version": "0.1",
						"channel": "regular",
					},
					map[string]interface{}{
						"version": "0.2",
						"channel": "fast",
					},
				},
			},
		},
	}
)

func TestList(t *testing.T) {
	t.Run("list modules from cluster", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(kyma.GVRModuleTemplate.GroupVersion())
		scheme.AddKnownTypes(kyma.GVRModuleReleaseMeta.GroupVersion())
		dynamicClient := dynamic_fake.NewSimpleDynamicClient(scheme,
			&testModuleTemplate1,
			&testModuleTemplate2,
			&testModuleTemplate3,
			&testModuleTemplate4,
			&testReleaseMeta1,
			&testReleaseMeta2,
		)

		modules, err := List(context.Background(), kyma.NewClient(dynamicClient))

		require.NoError(t, err)
		require.Equal(t, ModulesList(fixModuleList()), modules)
	})
}

func fixModuleList() []Module {
	return []Module{
		{
			Name: "keda",
			Versions: []ModuleVersion{
				{
					Repository: "url-3",
					Version:    "0.1",
					Channel:    "regular",
				},
				{
					Version: "0.2",
					Channel: "fast",
				},
			},
		},
		{
			Name: "serverless",
			Versions: []ModuleVersion{
				{
					Repository: "url-1",
					Version:    "0.0.1",
					Channel:    "fast",
				},
				{
					Repository: "url-2",
					Version:    "0.0.2",
				},
			},
		},
	}
}
