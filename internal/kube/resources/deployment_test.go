package resources

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_BuildDeployment(t *testing.T) {
	t.Parallel()
	t.Run("SERVICE_BINDING_ROOT env var behavior", func(t *testing.T) {
		t.Run("should not add SERVICE_BINDING_ROOT when no service binding secrets", func(t *testing.T) {
			deployment := buildDeployment(&CreateDeploymentOpts{
				Name:                       "test-app",
				Namespace:                  "default",
				Image:                      "test:image",
				ServiceBindingSecretMounts: types.ServiceBindingSecretArray{}, // empty
				Envs:                       []corev1.EnvVar{},
			})

			envVars := deployment.Spec.Template.Spec.Containers[0].Env
			for _, env := range envVars {
				require.NotEqual(t, "SERVICE_BINDING_ROOT", env.Name,
					"SERVICE_BINDING_ROOT should not be present when no service binding secrets are used")
			}
		})

		t.Run("should add SERVICE_BINDING_ROOT when service binding secrets are used", func(t *testing.T) {
			serviceBindingSecrets := types.ServiceBindingSecretArray{}
			_ = serviceBindingSecrets.Set("my-service-binding-secret")

			deployment := buildDeployment(&CreateDeploymentOpts{
				Name:                       "test-app",
				Namespace:                  "default",
				Image:                      "test:image",
				ServiceBindingSecretMounts: serviceBindingSecrets,
				Envs:                       []corev1.EnvVar{},
			})

			envVars := deployment.Spec.Template.Spec.Containers[0].Env
			found := false
			for _, env := range envVars {
				if env.Name == "SERVICE_BINDING_ROOT" {
					found = true
					require.Equal(t, "/bindings", env.Value)
					break
				}
			}
			require.True(t, found, "SERVICE_BINDING_ROOT should be present when service binding secrets are used")
		})
	})
}

func Test_ApplyDeployment(t *testing.T) {
	t.Parallel()
	t.Run("apply creates deployment when it does not exist", func(t *testing.T) {
		rdClient := &kube_fake.RootlessDynamicClient{}
		kubeClient := &kube_fake.KubeClient{
			TestRootlessDynamicInterface: rdClient,
		}

		err := ApplyDeployment(context.Background(), kubeClient, CreateDeploymentOpts{
			Name:      "test-app",
			Namespace: "default",
			Image:     "test:v1",
		})
		require.NoError(t, err)
		require.Len(t, rdClient.ApplyObjs, 1)
		require.Equal(t, "apps/v1", rdClient.ApplyObjs[0].GetAPIVersion())
		require.Equal(t, "Deployment", rdClient.ApplyObjs[0].GetKind())
		require.Equal(t, "test-app", rdClient.ApplyObjs[0].GetName())
		require.Equal(t, "default", rdClient.ApplyObjs[0].GetNamespace())
	})

	t.Run("apply updates deployment when it already exists", func(t *testing.T) {
		rdClient := &kube_fake.RootlessDynamicClient{}
		kubeClient := &kube_fake.KubeClient{
			TestRootlessDynamicInterface: rdClient,
		}

		// First apply
		err := ApplyDeployment(context.Background(), kubeClient, CreateDeploymentOpts{
			Name:      "test-app",
			Namespace: "default",
			Image:     "test:v1",
		})
		require.NoError(t, err)

		// Second apply with different image — no error
		err = ApplyDeployment(context.Background(), kubeClient, CreateDeploymentOpts{
			Name:      "test-app",
			Namespace: "default",
			Image:     "test:v2",
		})
		require.NoError(t, err)
		require.Len(t, rdClient.ApplyObjs, 2)

		// Verify the second apply carried the updated image
		containers, _, _ := unstructured.NestedSlice(rdClient.ApplyObjs[1].Object, "spec", "template", "spec", "containers")
		require.Len(t, containers, 1)
		container := containers[0].(map[string]any)
		require.Equal(t, "test:v2", container["image"])
	})
}
