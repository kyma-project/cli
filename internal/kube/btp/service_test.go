package btp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
)

func Test_btpClient_GetServiceInstance(t *testing.T) {
	t.Run("get ServiceInstance", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceInstance.GroupVersion())
		givenInstance := fixServiceInstance()
		btpClient := NewClient(
			dynamic_fake.NewSimpleDynamicClient(scheme, givenInstance),
		)

		expectedInstance := &ServiceInstance{}
		toStructured(t, givenInstance, expectedInstance)

		serviceInstance, err := btpClient.GetServiceInstance(context.Background(), "test-namespace", "test-name")
		require.NoError(t, err)
		require.Equal(t, expectedInstance, serviceInstance)
	})

	t.Run("not found error", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceInstance.GroupVersion())
		btpClient := NewClient(
			dynamic_fake.NewSimpleDynamicClient(scheme),
		)

		serviceInstance, err := btpClient.GetServiceInstance(context.Background(), "test-namespace", "test-name")
		require.ErrorContains(t, err, "serviceinstances.services.cloud.sap.com \"test-name\" not found")
		require.Nil(t, serviceInstance)
	})
}

func Test_btpClient_CreateServiceInstance(t *testing.T) {
	t.Run("create ServiceInstance", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceInstance.GroupVersion())
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme)
		btpClient := NewClient(
			dynamic,
		)

		expectedInstance := fixServiceInstance()

		givenInstance := &ServiceInstance{}
		toStructured(t, expectedInstance, givenInstance)

		err := btpClient.CreateServiceInstance(context.Background(), givenInstance)
		require.NoError(t, err)

		serviceInstance, err := dynamic.Resource(GVRServiceInstance).
			Namespace("test-namespace").
			Get(context.Background(), "test-name", v1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, expectedInstance, serviceInstance)
	})

	t.Run("already exists error", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceInstance.GroupVersion())
		givenInstance := fixServiceInstance()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, givenInstance)
		btpClient := NewClient(
			dynamic,
		)

		expectedInstance := &ServiceInstance{}
		toStructured(t, givenInstance, expectedInstance)

		err := btpClient.CreateServiceInstance(context.Background(), expectedInstance)
		require.ErrorContains(t, err, "serviceinstances.services.cloud.sap.com \"test-name\" already exists")
	})

	t.Run("converter error", func(t *testing.T) {
		btpClient := NewClient(nil)

		instance := &ServiceInstance{}
		toStructured(t, fixServiceInstance(), instance)

		// add func parameter that is highly not supported and will cause error
		instance.Spec.Parameters = func() {}

		err := btpClient.CreateServiceInstance(context.Background(), instance)
		require.ErrorContains(t, err, "unrecognized type: func")
	})
}

func Test_btpClient_GetServiceBinding(t *testing.T) {
	t.Run("get ServiceBinding", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceBinding.GroupVersion())
		givenBinding := fixServiceBinding()
		btpClient := NewClient(
			dynamic_fake.NewSimpleDynamicClient(scheme, givenBinding),
		)

		expectedBinding := &ServiceBinding{}
		toStructured(t, givenBinding, expectedBinding)

		serviceBinding, err := btpClient.GetServiceBinding(context.Background(), "test-namespace", "test-name")
		require.NoError(t, err)
		require.Equal(t, expectedBinding, serviceBinding)
	})

	t.Run("not found error", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceBinding.GroupVersion())
		btpClient := NewClient(
			dynamic_fake.NewSimpleDynamicClient(scheme),
		)

		serviceBinding, err := btpClient.GetServiceBinding(context.Background(), "test-namespace", "test-name")
		require.ErrorContains(t, err, "servicebindings.services.cloud.sap.com \"test-name\" not found")
		require.Nil(t, serviceBinding)
	})
}

func Test_btpClient_CreateServiceBinding(t *testing.T) {
	t.Run("create ServiceBinding", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceBinding.GroupVersion())
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme)
		btpClient := NewClient(
			dynamic,
		)

		givenBinding := fixServiceBinding()

		expectedBinding := &ServiceBinding{}
		toStructured(t, givenBinding, expectedBinding)

		err := btpClient.CreateServiceBinding(context.Background(), expectedBinding)
		require.NoError(t, err)

		serviceBinding, err := dynamic.Resource(GVRServiceBinding).
			Namespace("test-namespace").
			Get(context.Background(), "test-name", v1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, givenBinding, serviceBinding)
	})

	t.Run("already exists error", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceBinding.GroupVersion())
		givenBinding := fixServiceBinding()
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, givenBinding)
		btpClient := NewClient(
			dynamic,
		)

		expectedBinding := &ServiceBinding{}
		toStructured(t, givenBinding, expectedBinding)

		err := btpClient.CreateServiceBinding(context.Background(), expectedBinding)
		require.ErrorContains(t, err, "servicebindings.services.cloud.sap.com \"test-name\" already exists")
	})

	t.Run("converter error", func(t *testing.T) {
		btpClient := NewClient(nil)

		binding := &ServiceBinding{}
		toStructured(t, fixServiceBinding(), binding)

		// add func parameter that is highly not supported and will cause error
		binding.Spec.Parameters = func() {}

		err := btpClient.CreateServiceBinding(context.Background(), binding)
		require.ErrorContains(t, err, "unrecognized type: func")
	})
}

func Test_btpClient_IsBindingReadyFunc(t *testing.T) {
	t.Run("ready binding", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceBinding.GroupVersion())
		givenBinding := fixServiceBinding()
		givenBinding.Object["status"] = fixReadyCommonStatus()

		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, givenBinding)
		btpClient := NewClient(
			dynamic,
		)

		readyFn := btpClient.IsBindingReady(context.Background(), "test-namespace", "test-name")
		done, err := readyFn(context.Background())
		require.NoError(t, err)
		require.True(t, done)
	})

	t.Run("failed binding", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceBinding.GroupVersion())
		givenBinding := fixServiceBinding()
		givenBinding.Object["status"] = fixFailedCommonStatus()

		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, givenBinding)
		btpClient := NewClient(
			dynamic,
		)

		readyFn := btpClient.IsBindingReady(context.Background(), "test-namespace", "test-name")
		done, err := readyFn(context.Background())
		require.ErrorContains(t, err, "test message")
		require.False(t, done)
	})

	t.Run("binding not found", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceBinding.GroupVersion())

		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme)
		btpClient := NewClient(
			dynamic,
		)

		readyFn := btpClient.IsBindingReady(context.Background(), "test-namespace", "test-name")
		done, err := readyFn(context.Background())
		require.ErrorContains(t, err, "servicebindings.services.cloud.sap.com \"test-name\" not found")
		require.False(t, done)
	})
}

func Test_btpClient_IsInstanceReadyFunc(t *testing.T) {
	t.Run("ready instance", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceInstance.GroupVersion())
		givenInstance := fixServiceInstance()
		givenInstance.Object["status"] = fixReadyCommonStatus()

		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, givenInstance)
		btpClient := NewClient(
			dynamic,
		)

		readyFn := btpClient.IsInstanceReady(context.Background(), "test-namespace", "test-name")
		done, err := readyFn(context.Background())
		require.NoError(t, err)
		require.True(t, done)
	})

	t.Run("failed instance", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceInstance.GroupVersion())
		givenInstance := fixServiceInstance()
		givenInstance.Object["status"] = fixFailedCommonStatus()

		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, givenInstance)
		btpClient := NewClient(
			dynamic,
		)

		readyFn := btpClient.IsInstanceReady(context.Background(), "test-namespace", "test-name")
		done, err := readyFn(context.Background())
		require.ErrorContains(t, err, "test message")
		require.False(t, done)
	})

	t.Run("instance not found", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(GVRServiceInstance.GroupVersion())

		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme)
		btpClient := NewClient(
			dynamic,
		)

		readyFn := btpClient.IsInstanceReady(context.Background(), "test-namespace", "test-name")
		done, err := readyFn(context.Background())
		require.ErrorContains(t, err, "serviceinstances.services.cloud.sap.com \"test-name\" not found")
		require.False(t, done)
	})
}

func fixServiceInstance() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "services.cloud.sap.com/v1",
			"kind":       "ServiceInstance",
			"metadata": map[string]interface{}{
				"creationTimestamp": interface{}(nil),
				"name":              "test-name",
				"namespace":         "test-namespace",
			},
			"spec": map[string]interface{}{
				"externalName": "test",
			},
			"status": map[string]interface{}{},
		},
	}
}

func fixServiceBinding() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "services.cloud.sap.com/v1",
			"kind":       "ServiceBinding",
			"metadata": map[string]interface{}{
				"creationTimestamp": interface{}(nil),
				"name":              "test-name",
				"namespace":         "test-namespace",
			},
			"spec": map[string]interface{}{
				"externalName": "test",
			},
			"status": map[string]interface{}{},
		},
	}
}

func fixReadyCommonStatus() map[string]interface{} {
	return map[string]interface{}{
		"ready": "True",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "Succeeded",
				"status": "True",
			},
			map[string]interface{}{
				"type":   "Ready",
				"status": "True",
			},
		},
	}
}

func fixFailedCommonStatus() map[string]interface{} {
	return map[string]interface{}{
		"ready": "False",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":    "Failed",
				"status":  "True",
				"message": "test message",
			},
		},
	}
}

func toStructured(t *testing.T, u *unstructured.Unstructured, obj interface{}) {
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, obj)
	require.NoError(t, err)
}
