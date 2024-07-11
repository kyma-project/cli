package kyma

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube"
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

func TestUpdateDefaultKyma(t *testing.T) {
	type args struct {
		ctx    context.Context
		client kube.Client
		obj    *Kyma
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateDefaultKyma(tt.args.ctx, tt.args.client, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("UpdateDefaultKyma() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
