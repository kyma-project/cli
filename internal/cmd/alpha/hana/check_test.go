package hana

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube/btp"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
)

func Test_checkHanaInstance(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		testHana := fixTestHanaServiceInstance(name, namespace, nil)
		config := fixCheckConfig(name, namespace, testHana)
		err := checkHanaInstance(&config)
		require.Nil(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		testHana := fixTestHanaServiceInstance("other-name", namespace, nil)
		config := fixCheckConfig(name, namespace, testHana)
		err := checkHanaInstance(&config)
		require.NotNil(t, err)
		errMsg := err.String()
		require.Contains(t, errMsg, "failed to get resource data")
		require.Contains(t, errMsg, "serviceinstances.services.cloud.sap.com \"test-name\" not found")
		require.Contains(t, errMsg, "Make sure that Hana was provisioned")
	})
	t.Run("not ready", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		testHana := fixTestHanaServiceInstance(name, namespace, &map[string]interface{}{})
		config := fixCheckConfig(name, namespace, testHana)
		err := checkHanaInstance(&config)
		require.NotNil(t, err)
		errMsg := err.String()
		require.Contains(t, errMsg, "hana instance is not ready")
		require.Contains(t, errMsg, "Wait for provisioning of Hana resources")
		require.Contains(t, errMsg, "Check if Hana resources started without errors")
	})
}

func Test_checkHanaBinding(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		testHanaBinding := fixTestHanaServiceBinding(name, namespace, nil)
		config := fixCheckConfig(name, namespace, testHanaBinding)
		err := checkHanaBinding(&config)
		require.Nil(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		testHanaBinding := fixTestHanaServiceBinding("other-name", namespace, nil)
		config := fixCheckConfig(name, namespace, testHanaBinding)
		err := checkHanaBinding(&config)
		require.NotNil(t, err)
		errMsg := err.String()
		require.Contains(t, errMsg, "failed to get resource data")
		require.Contains(t, errMsg, "servicebindings.services.cloud.sap.com \"test-name\" not found")
		require.Contains(t, errMsg, "Make sure that Hana was provisioned")
	})
	t.Run("not ready", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		testHanaBinding := fixTestHanaServiceBinding(name, namespace, &map[string]interface{}{})
		config := fixCheckConfig(name, namespace, testHanaBinding)
		err := checkHanaBinding(&config)
		require.NotNil(t, err)
		errMsg := err.String()
		require.Contains(t, errMsg, "hana binding is not ready")
		require.Contains(t, errMsg, "Wait for provisioning of Hana resources")
		require.Contains(t, errMsg, "Check if Hana resources started without errors")
	})
}

func Test_checkHanaBindingURL(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		name := "test-name"
		urlName := name + "-url"
		namespace := "test-namespace"
		testHanaBinding := fixTestHanaServiceBinding(urlName, namespace, nil)
		config := fixCheckConfig(name, namespace, testHanaBinding)
		err := checkHanaBindingURL(&config)
		require.Nil(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		testHanaBinding := fixTestHanaServiceBinding("other-name", namespace, nil)
		config := fixCheckConfig(name, namespace, testHanaBinding)
		err := checkHanaBindingURL(&config)
		require.NotNil(t, err)
		errMsg := err.String()
		require.Contains(t, errMsg, "failed to get resource data")
		require.Contains(t, errMsg, "servicebindings.services.cloud.sap.com \"test-name-url\" not found")
		require.Contains(t, errMsg, "Make sure that Hana was provisioned")
	})
	t.Run("not ready", func(t *testing.T) {
		name := "test-name"
		urlName := name + "-url"
		namespace := "test-namespace"
		testHanaBinding := fixTestHanaServiceBinding(urlName, namespace, &map[string]interface{}{})
		config := fixCheckConfig(name, namespace, testHanaBinding)
		err := checkHanaBindingURL(&config)
		require.NotNil(t, err)
		errMsg := err.String()
		require.Contains(t, errMsg, "hana URL binding is not ready")
		require.Contains(t, errMsg, "Wait for provisioning of Hana resources")
		require.Contains(t, errMsg, "Check if Hana resources started without errors")
	})
}

func fixCheckConfig(name string, namespace string, objects ...runtime.Object) hanaCheckConfig {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(btp.GVRServiceInstance.GroupVersion())
	dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, objects...)
	config := hanaCheckConfig{
		KymaConfig:       &cmdcommon.KymaConfig{Ctx: context.Background()},
		KubeClientConfig: cmdcommon.KubeClientConfig{KubeClient: &kube_fake.FakeKubeClient{TestDynamicInterface: dynamic}},
		name:             name,
		namespace:        namespace,
		timeout:          0,
	}
	return config
}

func fixTestHanaServiceInstance(name, namespace string, status *map[string]interface{}) *unstructured.Unstructured {
	return fixTestHanaService("ServiceInstance", name, namespace, status)
}

func fixTestHanaServiceBinding(name, namespace string, status *map[string]interface{}) *unstructured.Unstructured {
	return fixTestHanaService("ServiceBinding", name, namespace, status)
}

func fixTestHanaService(kind, name, namespace string, status *map[string]interface{}) *unstructured.Unstructured {
	if status == nil {
		status = &map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Succeeded",
					"status": string(metav1.ConditionTrue),
				},
				map[string]interface{}{
					"type":   "Ready",
					"status": string(metav1.ConditionTrue),
				},
			},
			"ready": "True",
		}
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "services.cloud.sap.com/v1",
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"status": *status,
		},
	}
}
