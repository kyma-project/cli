package registry

import (
	"context"
	"testing"

	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
)

func TestGetConfig(t *testing.T) {
	t.Run("Should return the RegistryConfig", func(t *testing.T) {
		// given
		testRegistrySvc := fixTestRegistrySvc()
		testRegistryPod := fixTestRegistryPod()
		testRegistrySecret := fixTestRegistrySecret()
		testDockerRegistry := fixTestDockerRegistry()

		client := k8s_fake.NewSimpleClientset(testRegistrySvc, testRegistryPod, testRegistrySecret)

		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(DockerRegistryGVR.GroupVersion(), testDockerRegistry)
		dynamic := dynamic_fake.NewSimpleDynamicClient(scheme, testDockerRegistry)

		expectedRegistryConfig := &RegistryConfig{
			SecretName: testRegistrySecret.GetName(),
			SecretData: &SecretData{
				DockerConfigJSON: string(testRegistrySecret.Data[".dockerconfigjson"]),
				Username:         string(testRegistrySecret.Data["username"]),
				Password:         string(testRegistrySecret.Data["password"]),
				PullRegAddr:      string(testRegistrySecret.Data["pullRegAddr"]),
				PushRegAddr:      string(testRegistrySecret.Data["pushRegAddr"]),
			},
			PodMeta: &RegistryPodMeta{
				Name:      "serverless-docker-registry-7d4d7b7b4f-7z5zv",
				Namespace: "test-namespace",
				Port:      "5000",
			},
		}

		kubeClient := &kube_fake.FakeKubeClient{
			TestKubernetesInterface: client,
			TestDynamicInterface:    dynamic,
		}

		// when
		config, err := GetConfig(context.Background(), kubeClient)

		// then
		require.Nil(t, err)
		require.Equal(t, expectedRegistryConfig, config)
	})
}

func Test_getRegistrySecretConfig(t *testing.T) {
	t.Run("Should return the RegistryConfig", func(t *testing.T) {
		// given
		testRegistrySecret := fixTestRegistrySecret()
		client := k8s_fake.NewSimpleClientset(testRegistrySecret)
		expectedRegistryConfig := &SecretData{
			DockerConfigJSON: string(testRegistrySecret.Data[".dockerconfigjson"]),
			Username:         string(testRegistrySecret.Data["username"]),
			Password:         string(testRegistrySecret.Data["password"]),
			PullRegAddr:      string(testRegistrySecret.Data["pullRegAddr"]),
			PushRegAddr:      string(testRegistrySecret.Data["pushRegAddr"]),
		}

		// when
		config, err := getRegistrySecretData(context.Background(), client, "test-secret", "test-namespace")

		// then
		require.NoError(t, err)
		require.Equal(t, expectedRegistryConfig, config)
	})

	t.Run("Should return an error when the secret does not exist", func(t *testing.T) {
		// given
		client := k8s_fake.NewSimpleClientset()
		expectedErrorMsg := "secrets \"test-secret\" not found"

		// when
		config, err := getRegistrySecretData(context.Background(), client, "test-secret", "test-namespace")

		// then
		require.Error(t, err)
		require.Nil(t, config)
		require.Contains(t, err.Error(), expectedErrorMsg)
	})
}

func Test_getWorkloadMeta(t *testing.T) {
	t.Run("Should return the RegistryPodMeta", func(t *testing.T) {
		// given
		client := k8s_fake.NewSimpleClientset(fixTestRegistrySvc(), fixTestRegistryPod())
		config := &SecretData{
			PushRegAddr: "serverless-docker-registry.test-namespace.svc.cluster.local:5000",
		}
		expectedRegistryPodMeta := &RegistryPodMeta{
			Name:      "serverless-docker-registry-7d4d7b7b4f-7z5zv",
			Namespace: "test-namespace",
			Port:      "5000",
		}

		// when
		meta, err := getWorkloadMeta(context.Background(), client, config)

		// then
		require.NoError(t, err)
		require.Equal(t, expectedRegistryPodMeta, meta)
	})
	t.Run("Should return an error when no pods exist", func(t *testing.T) {
		// given
		client := k8s_fake.NewSimpleClientset(fixTestRegistrySvc())
		config := &SecretData{
			PushRegAddr: "serverless-docker-registry.test-namespace.svc.cluster.local:5000",
		}

		// when
		meta, err := getWorkloadMeta(context.Background(), client, config)

		// then
		require.Error(t, err)
		require.Nil(t, meta)
		require.Contains(t, err.Error(), "no ready registry pod found")
	})
}

func fixTestRegistrySvc() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "serverless-docker-registry",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "serverless-docker-registry",
			},
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.FromString("5000"),
				},
			},
		},
	}
}

func fixTestRegistryPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "serverless-docker-registry-7d4d7b7b4f-7z5zv",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "serverless-docker-registry",
			},
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.ContainersReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
}

func fixTestRegistrySecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(`{"auths": {}}`),
			"username":          []byte("testUsername"),
			"password":          []byte("testPassword"),
			"pullRegAddr":       []byte("localhost:32137"),
			"pushRegAddr":       []byte("serverless-docker-registry.test-namespace.svc.cluster.local:5000"),
		},
	}
}

func fixTestDockerRegistry() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1alpha1",
			"kind":       "DockerRegistry",
			"metadata": map[string]interface{}{
				"name":      "test-docker-registry",
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"internalAccess": map[string]interface{}{
					"secretName": "test-secret",
				},
				"state": "Ready",
			},
		},
	}
}
