package kyma

import (
	"context"
	"testing"

	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
)

func TestGetDefaultKyma(t *testing.T) {
	t.Run("get Kyma from the cluster", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		kubeClient := &kube_fake.FakeKubeClient{
			TestDynamicInterface: dynamic_fake.NewSimpleDynamicClient(scheme, fixDefaultKyma()),
		}

		expectedKyma := &Kyma{
			TypeMeta: v1.TypeMeta{
				APIVersion: "operator.kyma-project.io/v1beta2",
				Kind:       "Kyma",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "default",
				Namespace: "kyma-system",
			},
			Spec: KymaSpec{
				Channel: "fast",
				Modules: []Module{
					{
						Name: "test-module",
					},
				},
			},
		}

		kyma, err := GetDefaultKyma(context.Background(), kubeClient)
		require.NoError(t, err)
		require.Equal(t, expectedKyma, kyma)
	})

	t.Run("kyma not found", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		kubeClient := &kube_fake.FakeKubeClient{
			TestDynamicInterface: dynamic_fake.NewSimpleDynamicClient(scheme),
		}

		kyma, err := GetDefaultKyma(context.Background(), kubeClient)
		require.ErrorContains(t, err, "not found")
		require.Nil(t, kyma)
	})
}

func TestUpdateDefaultKyma(t *testing.T) {
	t.Run("update kyma", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		kubeClient := &kube_fake.FakeKubeClient{
			TestDynamicInterface: dynamic_fake.NewSimpleDynamicClient(scheme, fixDefaultKyma()),
		}

		expectedKyma := &Kyma{
			TypeMeta: v1.TypeMeta{
				APIVersion: "operator.kyma-project.io/v1beta2",
				Kind:       "Kyma",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "default",
				Namespace: "kyma-system",
			},
			Spec: KymaSpec{
				Channel: "fast",
				Modules: []Module{
					{
						Name:    "test-module",
						Channel: "regular",
					},
					{
						Name: "another-test-module",
					},
				},
			},
		}

		err := UpdateDefaultKyma(context.Background(), kubeClient, expectedKyma)
		require.NoError(t, err)

		u, err := kubeClient.Dynamic().Resource(GVRKyma).Namespace("kyma-system").Get(context.Background(), "default", v1.GetOptions{})
		require.NoError(t, err)

		kyma := &Kyma{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, kyma)
		require.NoError(t, err)
		require.Equal(t, expectedKyma, kyma)
	})

	t.Run("kyma not found", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRKyma.GroupVersion())
		kubeClient := &kube_fake.FakeKubeClient{
			TestDynamicInterface: dynamic_fake.NewSimpleDynamicClient(scheme),
		}

		expectedKyma := &Kyma{
			TypeMeta: v1.TypeMeta{
				APIVersion: "operator.kyma-project.io/v1beta2",
				Kind:       "Kyma",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "default",
				Namespace: "kyma-system",
			},
			Spec: KymaSpec{
				Channel: "fast",
				Modules: []Module{
					{
						Name:    "test-module",
						Channel: "regular",
					},
					{
						Name: "another-test-module",
					},
				},
			},
		}

		err := UpdateDefaultKyma(context.Background(), kubeClient, expectedKyma)
		require.ErrorContains(t, err, "not found")
	})
}

func fixDefaultKyma() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1beta2",
			"kind":       "Kyma",
			"metadata": map[string]interface{}{
				"name":      "default",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"channel": "fast",
				"modules": []interface{}{
					map[string]interface{}{
						"name": "test-module",
					},
				},
			},
		},
	}
}
