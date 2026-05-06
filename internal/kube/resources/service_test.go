package resources

import (
	"context"
	"fmt"
	"testing"

	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/istio"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_CreateAPIRule(t *testing.T) {
	t.Run("create apiRule", func(t *testing.T) {
		ctx := context.Background()
		rootlessdynamic := &kube_fake.RootlessDynamicClient{}
		apiRuleName := "apiRule"
		namespace := "default"
		host := "example.com"
		port := uint32(80)

		err := CreateAPIRule(ctx, rootlessdynamic, apiRuleName, namespace, host, port)

		require.NoError(t, err)
		require.Equal(t, 1, len(rootlessdynamic.ApplyObjs))
		require.Equal(t, fixAPIRule(apiRuleName, namespace, host, port), rootlessdynamic.ApplyObjs[0])
	})
	t.Run("do not allow creating existing apiRule", func(t *testing.T) {
		ctx := context.Background()
		rootlessdynamic := &kube_fake.RootlessDynamicClient{
			ReturnErr: fmt.Errorf("already exists"),
		}
		apiRuleName := "existing"
		namespace := "default"
		domain := "example.com"
		port := uint32(80)
		err := CreateAPIRule(ctx, rootlessdynamic, apiRuleName, namespace, domain, port)
		require.Contains(t, err.Error(), "already exists")
	})
}

func fixAPIRule(apiRuleName, namespace, host string, port uint32) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "gateway.kyma-project.io/v2alpha1",
			"kind":       "APIRule",
			"metadata": map[string]interface{}{
				"name":      apiRuleName,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       apiRuleName,
					"app.kubernetes.io/created-by": "kyma-cli",
				},
			},
			"spec": map[string]interface{}{
				"hosts": []interface{}{
					host,
				},
				"gateway": fmt.Sprintf("%s/%s", istio.DefaultGatewayNamespace, istio.DefaultGatewayName),
				"rules": []interface{}{
					map[string]interface{}{
						"path":    "/*",
						"methods": []interface{}{"GET", "POST", "PUT", "DELETE", "PATCH"},
						"noAuth":  true,
					},
				},
				"service": map[string]interface{}{
					"name":      apiRuleName,
					"namespace": namespace,
					"port":      int64(port),
				},
			},
			"status": map[string]interface{}{"lastProcessedTime": interface{}(nil), "state": ""},
		},
	}
}

func Test_ApplyService(t *testing.T) {
	t.Parallel()
	t.Run("apply creates service when it does not exist", func(t *testing.T) {
		rdClient := &kube_fake.RootlessDynamicClient{}
		kubeClient := &kube_fake.KubeClient{
			TestRootlessDynamicInterface: rdClient,
		}

		err := ApplyService(context.Background(), kubeClient, "test-svc", "default", 8080)
		require.NoError(t, err)
		require.Len(t, rdClient.ApplyObjs, 1)
		require.Equal(t, "v1", rdClient.ApplyObjs[0].GetAPIVersion())
		require.Equal(t, "Service", rdClient.ApplyObjs[0].GetKind())
		require.Equal(t, "test-svc", rdClient.ApplyObjs[0].GetName())
		require.Equal(t, "default", rdClient.ApplyObjs[0].GetNamespace())
	})

	t.Run("apply updates service when it already exists", func(t *testing.T) {
		rdClient := &kube_fake.RootlessDynamicClient{}
		kubeClient := &kube_fake.KubeClient{
			TestRootlessDynamicInterface: rdClient,
		}

		// First apply
		err := ApplyService(context.Background(), kubeClient, "test-svc", "default", 8080)
		require.NoError(t, err)

		// Second apply with different port — no error
		err = ApplyService(context.Background(), kubeClient, "test-svc", "default", 9090)
		require.NoError(t, err)
		require.Len(t, rdClient.ApplyObjs, 2)

		// Verify the second apply carried the updated port
		ports, _, _ := unstructured.NestedSlice(rdClient.ApplyObjs[1].Object, "spec", "ports")
		require.Len(t, ports, 1)
		port := ports[0].(map[string]any)
		require.Equal(t, int64(9090), port["port"])
	})
}
