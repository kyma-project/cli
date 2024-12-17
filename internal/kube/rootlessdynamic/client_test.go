package rootlessdynamic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	apimachinery_errors "k8s.io/apimachinery/pkg/api/errors"
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
		require.ErrorContains(t, err, "failed to discover API resource using discovery client: resource 'Secret' in group '', and version 'v1' not registered on cluster")
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

		err := client.ApplyMany(ctx, []unstructured.Unstructured{
			*clusterRole, *secret,
		})
		require.ErrorContains(t, err, "the server could not find the requested resource, GroupVersion \"v1\" not found")

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

func Test_Get(t *testing.T) {
	t.Run("get namespaced resource", func(t *testing.T) {
		obj, apiResource := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, obj)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		result, err := client.Get(ctx, obj)
		require.NoError(t, err)
		require.Equal(t, obj, result)
	})

	t.Run("get cluster-scoped resource", func(t *testing.T) {
		obj, apiResource := fixClusterRoleObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, obj)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		result, err := client.Get(ctx, obj)
		require.NoError(t, err)
		require.Equal(t, obj, result)
	})

	t.Run("get resource error because can't be discovered", func(t *testing.T) {
		obj, _ := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{
			{
				GroupVersion: "v1",
			},
		})

		_, err := client.Get(ctx, obj)
		require.ErrorContains(t, err, "failed to discover API resource using discovery client: resource 'Secret' in group '', and version 'v1' not registered on cluster")
	})
}

var (
	secretListObject = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "SecretList",
		"metadata": map[string]interface{}{
			"continue":        "",
			"resourceVersion": "",
		},
	}

	clusterRoleListObject = map[string]interface{}{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "ClusterRoleList",
		"metadata": map[string]interface{}{
			"continue":        "",
			"resourceVersion": "",
		},
	}
)

func Test_List(t *testing.T) {
	t.Run("list namespaced resource", func(t *testing.T) {
		obj, apiResource := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, obj)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		expectedResult := &unstructured.UnstructuredList{
			Object: secretListObject,
			Items:  []unstructured.Unstructured{*obj},
		}

		result, err := client.List(ctx, obj)
		require.NoError(t, err)
		require.Equal(t, expectedResult, result)
	})

	t.Run("list cluster-scoped resource", func(t *testing.T) {
		obj, apiResource := fixClusterRoleObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, obj)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		expectedResult := &unstructured.UnstructuredList{
			Object: clusterRoleListObject,
			Items:  []unstructured.Unstructured{*obj},
		}

		result, err := client.List(ctx, obj)
		require.NoError(t, err)
		require.Equal(t, expectedResult, result)
	})

	t.Run("get resource error because can't be discovered", func(t *testing.T) {
		obj, _ := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{
			{
				GroupVersion: "v1",
			},
		})

		_, err := client.List(ctx, obj)
		require.ErrorContains(t, err, "failed to discover API resource using discovery client: resource 'Secret' in group '', and version 'v1' not registered on cluster")
	})
}

func Test_Remove(t *testing.T) {
	t.Run("remove namespaced resource", func(t *testing.T) {
		obj, apiResource := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, obj)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		err := client.Remove(ctx, obj)
		require.Nil(t, err)

		_, getErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}).Namespace("kyma-system").Get(ctx, "test", metav1.GetOptions{})
		require.Error(t, getErr)
		require.ErrorContains(t, getErr, "secrets \"test\" not found")
	})

	t.Run("remove cluster-scoped resource", func(t *testing.T) {
		obj, apiResource := fixClusterRoleObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, obj)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{apiResource})

		err := client.Remove(ctx, obj)
		require.Nil(t, err)

		_, getErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		}).Get(ctx, "test", metav1.GetOptions{})
		require.Error(t, getErr)
		require.ErrorContains(t, getErr, "clusterroles.rbac.authorization.k8s.io \"test\" not found")
	})

	t.Run("remove namespaced resource error because can't be discovered", func(t *testing.T) {
		obj, _ := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{
			{
				GroupVersion: "v1",
			},
		})

		err := client.Remove(ctx, obj)
		require.ErrorContains(t, err, "failed to discover API resource using discovery client: resource 'Secret' in group '', and version 'v1' not registered on cluster")
	})
}

func Test_RemoveMany(t *testing.T) {
	t.Run("Remove many resources", func(t *testing.T) {
		clusterRole, clusterRoleApiResource := fixClusterRoleObjectAndApiResource()
		secret, secretApiResource := fixSecretObjectAndApiResource()
		ctx := context.Background()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme.Scheme, clusterRole, secret)
		client := fixRootlessDynamic(dynamic, []*metav1.APIResourceList{
			clusterRoleApiResource, secretApiResource,
		})

		err := client.RemoveMany(ctx, []unstructured.Unstructured{
			*clusterRole, *secret,
		})
		require.Nil(t, err)

		_, getErr := dynamic.Resource(schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		}).Get(ctx, "test", metav1.GetOptions{})
		require.Error(t, getErr)
		require.ErrorContains(t, getErr, "clusterroles.rbac.authorization.k8s.io \"test\" not found")

		_, getErr = dynamic.Resource(schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}).Namespace("kyma-system").Get(ctx, "test", metav1.GetOptions{})
		require.Error(t, getErr)
		require.ErrorContains(t, getErr, "secrets \"test\" not found")
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

func fixRootlessDynamic(dynamic dynamic.Interface, apiResources []*metav1.APIResourceList) Interface {
	return NewClientWithApplyFunc(
		dynamic,
		&clientgo_fake.FakeDiscovery{
			Fake: &clientgo_testing.Fake{
				Resources: apiResources,
			},
		},
		fixApplyFunc,
	)
}

// this func is a testing version of the Apply func that can't be used in tests because of dynamic.FakeDynamicClient limitations
func fixApplyFunc(ctx context.Context, ri dynamic.ResourceInterface, u *unstructured.Unstructured) error {
	_, err := ri.Create(ctx, u, metav1.CreateOptions{
		FieldManager: "cli",
	})
	if apimachinery_errors.IsAlreadyExists(err) {
		_, err = ri.Update(ctx, u, metav1.UpdateOptions{
			FieldManager: "cli",
		})
	}

	return err
}
