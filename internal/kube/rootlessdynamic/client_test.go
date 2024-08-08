package rootlessdynamic

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgo_fake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clientgo_testing "k8s.io/client-go/testing"
)

func Test_Apply(t *testing.T) {
	t.Run("create namespaced resource", func(t *testing.T) {
		obj, apiResource := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		err := client.Apply(ctx, obj)
		require.Nil(t, err)

		clusterObj, clusterErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}).Namespace("kyma-system").Get(ctx, "test", metav1.GetOptions{})
		require.NoError(t, clusterErr)
		require.Equal(t, obj, clusterObj)
	})

	t.Run("create cluster-scope resource", func(t *testing.T) {
		obj, apiResource := fixClusterRoleObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		err := client.Apply(ctx, obj)
		require.Nil(t, err)

		clusterObj, clusterErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		}).Get(ctx, "test", metav1.GetOptions{})
		require.NoError(t, clusterErr)
		require.Equal(t, obj, clusterObj)
	})

	t.Run("update namespaced resource", func(t *testing.T) {
		obj, apiResource := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, obj)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		obj.Object["data"] = map[string]interface{}{
			"key": "value2",
		}

		err := client.Apply(ctx, obj)
		require.Nil(t, err)

		clusterObj, clusterErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}).Namespace("kyma-system").Get(ctx, "test", metav1.GetOptions{})
		require.NoError(t, clusterErr)
		require.Equal(t, obj, clusterObj)
	})

	t.Run("update cluster-scope resource", func(t *testing.T) {
		obj, apiResource := fixClusterRoleObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, obj)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		obj.Object["rules"] = map[string]interface{}{
			"apiGroups": []interface{}{
				"",
			},
		}

		err := client.Apply(ctx, obj)
		require.Nil(t, err)

		clusterObj, clusterErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		}).Get(ctx, "test", metav1.GetOptions{})
		require.NoError(t, clusterErr)
		require.Equal(t, obj, clusterObj)
	})

	t.Run("create namespaced resource error because can't be discovered", func(t *testing.T) {
		obj, _ := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{
			{
				GroupVersion: "v1",
			},
		})

		err := client.Apply(ctx, obj)
		expectErr := clierror.Wrap(
			errors.New("resource 'Secret' in group '', and version 'v1' not registered on cluster"),
			clierror.New("failed to discover API resource using discovery client"))
		require.Equal(t, expectErr, err)
	})
}

func Test_client_ApplyMany(t *testing.T) {
	t.Run("Apply many resources", func(t *testing.T) {
		clusterRole, clusterRoleApiResource := fixClusterRoleObjectAndApiResource()
		secret, secretApiResource := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{
			clusterRoleApiResource, secretApiResource,
		})

		err := client.ApplyMany(ctx, []unstructured.Unstructured{
			*clusterRole, *secret,
		})
		require.Nil(t, err)

		createdClusterRole, getErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		}).Get(ctx, "test", metav1.GetOptions{})
		require.NoError(t, getErr)
		require.Equal(t, clusterRole, createdClusterRole)

		createdSecret, getErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}).Namespace("kyma-system").Get(ctx, "test", metav1.GetOptions{})
		require.NoError(t, getErr)
		require.Equal(t, secret, createdSecret)
	})

	t.Run("failed to discover second resource", func(t *testing.T) {
		clusterRole, clusterRoleApiResource := fixClusterRoleObjectAndApiResource()
		secret, _ := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{
			clusterRoleApiResource,
		})

		expectedErr := clierror.Wrap(errors.New(
			"the server could not find the requested resource, GroupVersion \"v1\" not found"),
			clierror.New("failed to discover API resource using discovery client"))

		err := client.ApplyMany(ctx, []unstructured.Unstructured{
			*clusterRole, *secret,
		})
		require.Equal(t, expectedErr, err)

		createdClusterRole, getErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		}).Get(ctx, "test", metav1.GetOptions{})
		require.NoError(t, getErr)
		require.Equal(t, clusterRole, createdClusterRole)

		createdSecret, getErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}).Namespace("kyma-system").Get(ctx, "test", metav1.GetOptions{})
		require.Error(t, getErr)
		require.Nil(t, createdSecret)
	})
}

func fixSecretObjectAndApiResource() (*unstructured.Unstructured, *metav1.APIResourceList) {
	return &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "kyma-system",
				},
				"data": map[string]interface{}{
					"key": "value",
				},
			},
		}, &metav1.APIResourceList{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{
					Group:        "",
					Version:      "v1",
					Kind:         "Secret",
					SingularName: "secret",
					Name:         "secrets",
					Namespaced:   true,
				},
			},
		}
}

func fixClusterRoleObjectAndApiResource() (*unstructured.Unstructured, *metav1.APIResourceList) {
	return &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "ClusterRole",
				"metadata": map[string]interface{}{
					"name": "test",
				},
			},
		}, &metav1.APIResourceList{
			GroupVersion: "rbac.authorization.k8s.io/v1",
			APIResources: []metav1.APIResource{
				{
					Group:        "rbac.authorization.k8s.io",
					Version:      "v1",
					Kind:         "ClusterRole",
					SingularName: "clusterrole",
					Name:         "clusterroles",
					Namespaced:   false,
				},
			},
		}
}

func fixRootlessDynamic(dynamic dynamic.Interface, apiResources []*metav1.APIResourceList) *client {
	return &client{
		dynamic: dynamic,
		discovery: &clientgo_fake.FakeDiscovery{
			Fake: &clientgo_testing.Fake{
				Resources: apiResources,
			},
		},
	}
}
